/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restorecontroller

import (
	"os"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"
)

const restoreControllerName = "CStorRestore"

// RestoreController is the controller implementation for CStorRestore resources.
type RestoreController struct {
	// kubeclientset is a standard kubernetes clientset.
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// RestoreSynced is used for caches sync to get populated
	RestoreSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewCStorRestoreController returns a new cStor restore controller instance
func NewCStorRestoreController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *RestoreController {

	// obtain references to shared index informers for the CStorRestore resources.
	CStorRestoreInformer := cStorInformerFactory.Openebs().V1alpha1().CStorRestores()

	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		glog.Errorf("Failed to add scheme :%v", err.Error())
		return nil
	}

	// Create event broadcaster
	// Add cStor-Replica-controller types to the default Kubernetes Scheme so Events can be
	// logged for cStor-Replica-controller types.
	glog.V(4).Info("Creating restore event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: restoreControllerName})

	controller := &RestoreController{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		RestoreSynced: CStorRestoreInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorRestore"),
		recorder:      recorder,
	}

	glog.Info("Setting up event handlers for restore")

	// Clean any pending restore for this cstor pool
	controller.cleanupOldRestore(clientset)

	// Instantiating QueueLoad before entering workqueue.
	q := common.QueueLoad{}

	// Set up an event handler for when cStorReplica resources change.
	CStorRestoreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// ToDo : Enqueue object for processing in case of added event
			// Note: AddFunc is called when a new object comes into etcd
			// Note : In case controller reboots and existing object in etcd can cause delivery of
			// add event when the controller comes again. Be careful in this part and handle accordingly.
			rst := obj.(*apis.CStorRestore)

			if !IsRightCStorPoolMgmt(rst) {
				return
			}

			controller.handleRSTAddEvent(rst, &q)
		},
		UpdateFunc: func(oldVar, newVar interface{}) {
			// Note : UpdateFunc is called in following three cases:
			// 1. When object is updated/patched i.e. Resource version of object changes.
			// 2. When object is deleted i.e. the deletion timestamp of object is set.
			// 3. After every resync interval.
			newrst := newVar.(*apis.CStorRestore)
			oldrst := oldVar.(*apis.CStorRestore)

			if !IsRightCStorPoolMgmt(newrst) {
				return
			}

			controller.handleRSTUpdateEvent(oldrst, newrst, &q)
		},
		DeleteFunc: func(obj interface{}) {
			rst := obj.(*apis.CStorRestore)
			if !IsRightCStorPoolMgmt(rst) {
				return
			}
			glog.Infof("Restore resource deleted event: %v, %v", rst.ObjectMeta.Name, string(rst.ObjectMeta.UID))

		},
	})
	return controller
}

// enqueueCStorRestore takes a CStorRestore resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorRestore.
func (c *RestoreController) enqueueCStorRestore(obj *apis.CStorRestore, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

// handleRSTAddEvent is to handle add operation of restore controller
func (c *RestoreController) handleRSTAddEvent(rst *apis.CStorRestore, q *common.QueueLoad) {
	q.Operation = common.QOpAdd
	glog.Infof("cStorRestore event added: %v, %v", rst.ObjectMeta.Name, string(rst.ObjectMeta.UID))
	c.recorder.Event(rst, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageCreateSynced))
	c.enqueueCStorRestore(rst, *q)
}

func (c *RestoreController) handleRSTUpdateEvent(oldrst, newrst *apis.CStorRestore, q *common.QueueLoad) {
	glog.Infof("Received Update for restore:%s", oldrst.ObjectMeta.Name)

	// If there is no change in status then we will ignore the event
	if newrst.Status == oldrst.Status {
		return
	}

	if IsDoneStatus(newrst) || IsFailedStatus(newrst) {
		return
	}

	if IsDestroyEvent(newrst) {
		q.Operation = common.QOpDestroy
		glog.Infof("cStorRestore Destroy event : %v, %v", newrst.ObjectMeta.Name, string(newrst.ObjectMeta.UID))
		c.recorder.Event(newrst, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageDestroySynced))
	} else {
		glog.Infof("cStorRestore Modify event : %v, %v", newrst.ObjectMeta.Name, string(newrst.ObjectMeta.UID))

		q.Operation = common.QOpSync
		c.recorder.Event(newrst, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageModifySynced))
		glog.Infof("Done modify event %v", newrst.Name)
	}
	c.enqueueCStorRestore(newrst, *q)
}

// cleanupOldRestore set fail status to old pending restore
func (c *RestoreController) cleanupOldRestore(clientset clientset.Interface) {
	rstlabel := "cstorpool.openebs.io/uid=" + os.Getenv(string(common.OpenEBSIOCStorID))
	rstlistop := metav1.ListOptions{
		LabelSelector: rstlabel,
	}
	rstlist, err := clientset.OpenebsV1alpha1().CStorRestores(metav1.NamespaceAll).List(rstlistop)
	if err != nil {
		return
	}

	for _, rst := range rstlist.Items {
		switch rst.Status {
		case apis.RSTCStorStatusDone:
			continue
		default:
			//Set restore status as failed
			updateRestoreStatus(clientset, rst, apis.RSTCStorStatusFailed)
		}
	}
}

// updateRestoreStatus will update the restore status to given status
func updateRestoreStatus(clientset clientset.Interface, rst apis.CStorRestore, status apis.CStorRestoreStatus) {
	rst.Status = status

	_, err := clientset.OpenebsV1alpha1().CStorRestores(rst.Namespace).Update(&rst)
	if err != nil {
		glog.Errorf("Failed to update restore(%s) status(%s)", status, rst.Name)
		return
	}
}

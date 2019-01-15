/*
Copyright 2018 The OpenEBS Authors.

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
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

	//clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	//openebsScheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/scheme"
	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"
)

const restoreControllerName = "CStorRestore"

// CStorRestoreController is the controller implementation for CStorRestore resources.
type CStorRestoreController struct {
	// kubeclientset is a standard kubernetes clientset.
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// CStorRestoreSynced is used for caches sync to get populated
	CStorRestoreSynced cache.InformerSynced

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

// NewCStorRestoreController returns a new cStor Replica controller instance
func NewCStorRestoreController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorRestoreController {

	// obtain references to shared index informers for the CStorRestore resources.
	CStorRestoreInformer := cStorInformerFactory.Openebs().V1alpha1().CStorRestores()

	openebsScheme.AddToScheme(scheme.Scheme)

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

	controller := &CStorRestoreController{
		kubeclientset:      kubeclientset,
		clientset:          clientset,
		CStorRestoreSynced: CStorRestoreInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorRestore"),
		recorder:           recorder,
	}

	glog.Info("Setting up event handlers for restore")

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

			if IsDeletionFailedBefore(rst) || IsErrorDuplicate(rst) {
				return
			}
			q.Operation = common.QOpAdd
			glog.Infof("cStorRestore Added event : %v, %v", rst.ObjectMeta.Name, string(rst.ObjectMeta.UID))
			controller.recorder.Event(rst, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageCreateSynced))
			rst.Status.Phase = apis.RSTStatusPending
			rst, err := controller.clientset.OpenebsV1alpha1().CStorRestores(rst.Namespace).Update(rst)
			if err != nil {
				glog.Errorf("Unable to update cstor restore cr: %v", err)
				return
			}
			rstData := create_rst_data(rst)

			_, err = controller.clientset.OpenebsV1alpha1().CStorRestoreDatas(rst.Namespace).Create(rstData)
			if err != nil {
				glog.Errorf("Failed to create restoredata: error '%s'", err.Error())
				return
			}
			controller.enqueueCStorRestore(rst, q)
		},
		UpdateFunc: func(old, new interface{}) {
			//controller.enqueueCStorRestore(newRST, q)
			// ToDo : Enqueue object for processing in case of update event
			// Note : UpdateFunc is called in following three cases:
			// 1. When object is updated/patched i.e. Resource version of object changes.
			// 2. When object is deleted i.e. the deletion timestap of object is set.
			// 3. After every resync interval.
			newRST := new.(*apis.CStorRestore)
			oldRST := old.(*apis.CStorRestore)
			glog.Infof("Received Update %s", newRST.Spec.Name)
			if !IsRightCStorPoolMgmt(newRST) {
				return
			}
			if IsOnlyStatusChange(oldRST, newRST) {
				glog.Infof("Only cSB status change: %v, %v", newRST.ObjectMeta.Name, string(newRST.ObjectMeta.UID))
				return
			}
			if IsDeletionFailedBefore(newRST) || IsErrorDuplicate(newRST) {
				return
			}
			if newRST.ResourceVersion == oldRST.ResourceVersion {
				q.Operation = common.QOpSync
				glog.Infof("Received CstorRestore status sync event for %s", newRST.ObjectMeta.Name)
				controller.recorder.Event(newRST, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.StatusSynced))
			} else if IsDestroyEvent(newRST) {
				q.Operation = common.QOpDestroy
				glog.Infof("cStorRestore Destroy event : %v, %v", newRST.ObjectMeta.Name, string(newRST.ObjectMeta.UID))
				controller.recorder.Event(newRST, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageDestroySynced))
			} else {
				glog.Infof("cStorRestore Modify event : %v, %v", newRST.ObjectMeta.Name, string(newRST.ObjectMeta.UID))

				q.Operation = common.QOpModify
				listOptions := v1.ListOptions{}
				var restoreData *apis.CStorRestoreData
				restoreDataList, _ := controller.clientset.OpenebsV1alpha1().CStorRestoreDatas(newRST.Namespace).List(listOptions)
				for _, rstData := range restoreDataList.Items {
					if rstData.Name == newRST.Name {
						restoreData = &rstData
					}
				}
				if restoreData == nil {
					glog.Infof("Failed to find restoreData for %s", newRST.Spec.Name)
					rstData := create_rst_data(newRST)
					_, err := controller.clientset.OpenebsV1alpha1().CStorRestoreDatas(newRST.Namespace).Create(rstData)
					if err != nil {
						glog.Errorf("Failed to create restoredata: error '%s'", err.Error())
						return
					}
					glog.Infof("Successfully Created restoreData for %s", newRST.Spec.Name)
				} else {
					/*
						update_rst_data(newRST, restoreData)
						restoreData, err = controller.clientset.OpenebsV1alpha1().CStorRestoreDatas(newRST.Namespace).Update(restoreData)
						if err != nil {
							glog.Errorf("Failed to update restoredata: error '%s'", err.Error())
						}
					*/
				}
				controller.recorder.Event(newRST, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageModifySynced))
				glog.Infof("Done modify event %v", newRST.Name)
			}
			controller.enqueueCStorRestore(newRST, q)
		},
		DeleteFunc: func(obj interface{}) {
			rst := obj.(*apis.CStorRestore)
			if !IsRightCStorPoolMgmt(rst) {
				return
			}
			glog.Infof("rst Resource deleted event: %v, %v", rst.ObjectMeta.Name, string(rst.ObjectMeta.UID))

		},
	})
	return controller
}

// enqueueCStorRestore takes a CStorRestore resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorRestore.
func (c *CStorRestoreController) enqueueCStorRestore(obj *apis.CStorRestore, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

func create_rst_data(rst *apis.CStorRestore) *apis.CStorRestoreData {
	restoreData := &apis.CStorRestoreData{}
	restoreData.Name = rst.Name
	restoreData.Namespace = rst.Namespace
	restoreData.Spec.Name = rst.Spec.Name
	restoreData.Spec.VolumeName = rst.Spec.VolumeName
	return restoreData
}

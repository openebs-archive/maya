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

package replicacontroller

import (
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
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
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
)

const replicaControllerName = "CStorVolumeReplica"

// CStorVolumeReplicaController is the controller implementation for cStorVolumeReplica resources.
type CStorVolumeReplicaController struct {
	// kubeclientset is a standard kubernetes clientset.
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// cStorReplicaSynced is used for caches sync to get populated
	cStorReplicaSynced cache.InformerSynced

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

// NewCStorVolumeReplicaController returns a new cStor Replica controller instance
func NewCStorVolumeReplicaController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorVolumeReplicaController {

	// obtain references to shared index informers for the cStorReplica resources.
	cStorReplicaInformer := cStorInformerFactory.Openebs().V1alpha1().CStorVolumeReplicas()

	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		glog.Errorf("failed to add to scheme: error {%v}", err)
	}
	// Create event broadcaster
	// Add cStor-Replica-controller types to the default Kubernetes Scheme so Events can be
	// logged for cStor-Replica-controller types.
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: replicaControllerName})

	controller := &CStorVolumeReplicaController{
		kubeclientset:      kubeclientset,
		clientset:          clientset,
		cStorReplicaSynced: cStorReplicaInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorVolumeReplica"),
		recorder:           recorder,
	}

	glog.Info("Setting up event handlers")

	// Instantiating QueueLoad before entering workqueue.
	q := common.QueueLoad{}

	// Set up an event handler for when cStorReplica resources change.
	cStorReplicaInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cVR := obj.(*apis.CStorVolumeReplica)
			if !IsRightCStorVolumeReplica(cVR) {
				return
			}
			if IsDeletionFailedBefore(cVR) || IsErrorDuplicate(cVR) {
				return
			}
			q.Operation = common.QOpAdd
			glog.Infof("cStorVolumeReplica Added event : %v, %v", cVR.ObjectMeta.Name, string(cVR.ObjectMeta.UID))
			controller.recorder.Event(cVR, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageCreateSynced))

			// For New request phase of cVR will be empty
			// ToDO: Need to have an annotation in CSP and CVR which will state
			// about recreation events.
			// For every restart of the cstor-pool-mgmt container replica
			// watcher will get add event
			if IsEmptyStatus(cVR) {
				cVR.Status.Phase = apis.CVRStatusInit
			} else {
				cVR.Status.Phase = apis.CVRStatusRecreate
			}

			cVR, _ = controller.clientset.OpenebsV1alpha1().CStorVolumeReplicas(cVR.Namespace).Update(cVR)
			controller.enqueueCStorReplica(cVR, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newCVR := new.(*apis.CStorVolumeReplica)
			oldCVR := old.(*apis.CStorVolumeReplica)
			if !IsRightCStorVolumeReplica(newCVR) {
				return
			}
			if IsOnlyStatusChange(oldCVR, newCVR) {
				glog.Infof("Only cVR status change: %v, %v", newCVR.ObjectMeta.Name, string(newCVR.ObjectMeta.UID))
				return
			}
			// ToDo: Need to have statuses in more organised manner
			// ToDo: IsErrorDuplicate(newCVR) is ignored as of now.

			// If DeletionFails -- cVR will not be reconciled.
			// Anyway -- due to removal of finalizers this cVR will not exist for reconciliation. But this code
			// is present if we might need to add finalizers.
			if IsDeletionFailedBefore(newCVR) {
				return
			}
			// Periodic resync will send update events for all known cStorReplica.
			// Two different versions of the same cStorReplica will always have different RVs.
			if newCVR.ResourceVersion == oldCVR.ResourceVersion {
				q.Operation = common.QOpSync
				glog.Infof("CstorVolumeReplica status sync event for %s", newCVR.ObjectMeta.Name)
				controller.recorder.Event(newCVR, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.StatusSynced))
			} else if IsDestroyEvent(newCVR) {
				q.Operation = common.QOpDestroy
				glog.Infof("cStorVolumeReplica Destroy event : %v, %v", newCVR.ObjectMeta.Name, string(newCVR.ObjectMeta.UID))
				controller.recorder.Event(newCVR, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageDestroySynced))
			} else {
				q.Operation = common.QOpModify
				glog.Infof("cStorVolumeReplica Modify event : %v, %v", newCVR.ObjectMeta.Name, string(newCVR.ObjectMeta.UID))
				controller.recorder.Event(newCVR, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageModifySynced))
				return // will be removed once modify is implemented
			}
			controller.enqueueCStorReplica(newCVR, q)
		},
		DeleteFunc: func(obj interface{}) {
			cVR := obj.(*apis.CStorVolumeReplica)
			if !IsRightCStorVolumeReplica(cVR) {
				return
			}
			glog.Infof("cVR Resource deleted event: %v, %v", cVR.ObjectMeta.Name, string(cVR.ObjectMeta.UID))
		},
	})

	return controller
}

// enqueueCStorReplica takes a CStorReplica resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorReplica.
func (c *CStorVolumeReplicaController) enqueueCStorReplica(obj *apis.CStorVolumeReplica, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

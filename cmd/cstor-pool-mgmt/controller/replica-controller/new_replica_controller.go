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

// CStorVolumeReplicaController is the controller
// for CVR resources.
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

// NewCStorVolumeReplicaController returns a new instance
// of CVR controller
func NewCStorVolumeReplicaController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory,
) *CStorVolumeReplicaController {

	// obtain references to shared index informers
	// for CVR resources.
	cvrInformer := cStorInformerFactory.
		Openebs().
		V1alpha1().
		CStorVolumeReplicas()

	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		glog.Errorf("failed to initialise cvr controller: %v", err)
	}

	// add cvr controller types to default Kubernetes scheme
	// to enable logging of cvr contrller events
	glog.V(4).Info("creating event broadcaster for cvr")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// start sending events received from this event broadcaster
	// to the assigned event handler
	//
	// Its return value can be ignored or used to stop recording, if
	// desired.
	//
	// Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kubeclientset.CoreV1().Events(""),
		},
	)
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		corev1.EventSource{Component: replicaControllerName},
	)

	controller := &CStorVolumeReplicaController{
		kubeclientset:      kubeclientset,
		clientset:          clientset,
		cStorReplicaSynced: cvrInformer.Informer().HasSynced,
		workqueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"CStorVolumeReplica",
		),
		recorder: recorder,
	}

	glog.Info("will set up informer event handlers for cvr")

	ql := common.QueueLoad{}

	cvrInformer.Informer().
		AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cvrObj := obj.(*apis.CStorVolumeReplica)

				if !IsRightCStorVolumeReplica(cvrObj) || IsErrorDuplicate(cvrObj) {
					// do nothing
					return
				}

				glog.V(4).Infof(
					"received informer add event for cvr {%s}",
					cvrObj.Name,
				)

				controller.recorder.Event(
					cvrObj,
					corev1.EventTypeNormal,
					string(common.SuccessSynced),
					string(common.MessageCreateSynced),
				)

				// new cvr requests will have phase as blank
				//
				// NOTE:
				//  for every restart of controller container
				// this informer handler will get add event
				// for each cvr resource present in k8s
				if IsEmptyStatus(cvrObj) {
					cvrObj.Status.Phase = apis.CVRStatusInit
				} else {
					cvrObj.Status.Phase = apis.CVRStatusRecreate
				}

				cvrObj, _ = controller.clientset.
					OpenebsV1alpha1().
					CStorVolumeReplicas(cvrObj.Namespace).
					Update(cvrObj)

					// push this operation to workqueue
				ql.Operation = common.QOpAdd
				controller.enqueueCStorReplica(cvrObj, ql)
			},

			// TODO @amitkumardas
			//
			// Need to think of writing more manageable code
			// In the current code, ordering of conditions
			// is very important. I am sure these conditions
			// will only increase as we release more versions.
			//
			// This logic has tried to handle multiple
			// responsibilities. IMO this particular informer
			// handler should act only as a **switch** to continue
			// to handle this change further or just reject this
			// change.
			//
			// We need to do a good job to categorise these
			// **if** conditions. Only the reject related conditions
			// should be here. Other conditions should be part
			// of actual business logic that handles change to
			// a resource.
			UpdateFunc: func(old, new interface{}) {
				newCVR := new.(*apis.CStorVolumeReplica)
				oldCVR := old.(*apis.CStorVolumeReplica)

				if !IsRightCStorVolumeReplica(newCVR) {
					// do nothing
					return
				}

				glog.V(4).Infof(
					"received informer update event for cvr {%s}",
					newCVR.Name,
				)

				if IsDestroyEvent(newCVR) {
					controller.recorder.Event(
						newCVR,
						corev1.EventTypeNormal,
						string(common.SuccessSynced),
						string(common.MessageDestroySynced),
					)

					// push this operation to workqueue
					ql.Operation = common.QOpDestroy
					controller.enqueueCStorReplica(newCVR, ql)
					return
				}

				if IsErrorDuplicate(newCVR) || IsOnlyStatusChange(oldCVR, newCVR) {
					// do nothing
					return
				}

				if newCVR.ResourceVersion != oldCVR.ResourceVersion {
					// cvr modify is not implemented
					// hence below is commented
					// ql.Operation = common.QOpModify

					controller.recorder.Event(
						newCVR,
						corev1.EventTypeNormal,
						string(common.SuccessSynced),
						string(common.MessageModifySynced),
					)

					// no further handling needed
					return
				}

				// finally !!!
				controller.recorder.Event(
					newCVR,
					corev1.EventTypeNormal,
					string(common.SuccessSynced),
					string(common.StatusSynced),
				)

				// push this operation to workqueue
				ql.Operation = common.QOpSync
				controller.enqueueCStorReplica(newCVR, ql)
			},

			DeleteFunc: func(obj interface{}) {
				cvrObj := obj.(*apis.CStorVolumeReplica)

				glog.V(4).Infof(
					"received informer delete event for cvr {%s}",
					cvrObj.Name,
				)

				// this is a noop since cvr delete is
				// handled in UpdateFunc
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

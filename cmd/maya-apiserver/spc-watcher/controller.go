/*
Copyright 2017 The OpenEBS Authors

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

package spc

import (
	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "spc-controller"
const (
	addEvent    = "add"
	updateEvent = "update"
	deleteEvent = "delete"
	ignoreEvent = "ignore"
)

// Controller is the controller implementation for SPC resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// spcSynced is used for caches sync to get populated
	spcSynced cache.InformerSynced

	// deletedIndexer holds deleted resource to be retreived after workqueue
	deletedIndexer cache.Indexer

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// queueLoad is the object or load that will be pushed into the
	// workqueue for later retrieval and processing.
	queueLoad QueueLoad
}

// NewController returns a new controller
func NewController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	spcInformerFactory informers.SharedInformerFactory) *Controller {
	// obtain references to shared index informers for the SPC resources
	spcInformer := spcInformerFactory.Openebs().V1alpha1().StoragePoolClaims()
	// Create event broadcaster
	// Add new-controller types to the default Kubernetes Scheme so Events can be
	// logged for new-controller types.
	openebsScheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	queueLoad := QueueLoad{}
	controller := &Controller{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		deletedIndexer: cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc,
			cache.Indexers{}),
		spcSynced: spcInformer.Informer().HasSynced,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SPC"),
		recorder:  recorder,
		queueLoad: queueLoad,
	}

	glog.Info("Setting up event handlers")

	// Set up an event handler for when SPC resources change
	spcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addSpc,

		// Informer will send update event along with object in following cases:
		// 1. In case the object is updated ( Change of Resource Version)

		// 2. In case the object is deleted by using a finalizer.
		//    Some of the kubectl version e.g. v1.11.0 will make use of a finalizer beforehand deletion.
		//    Hence in this case the delete of an object will cause a invoke of UpdateFunc
		//    as the finalizer addition will cause Resource Version to change.
		//
		//    But some of the kubectl version will not make use of finalizer for
		//    object deletion.
		//    Whatever be the case, delete of the resource will always invoke DeleteFunc
		//    and hence the delete event will be captured by it and the delete event
		//    that will be delivered in UpdateFunc due to specific kubectl versions will
		//    be suppressed.(see in updateSpc function, where delete event is marked as ignoreEvent)

		// 3. After every fixed amount of time which is know as reSync Period.
		//    ReSync period can be set to values we want. It can help in reconiciliation.
		UpdateFunc: controller.updateSpc,

		DeleteFunc: controller.deleteSpc,
	})

	return controller
}

func (c *Controller) addSpc(obj interface{}) {
	spcObject := obj.(*apis.StoragePoolClaim)
	c.queueLoad.Operation = addEvent
	c.queueLoad.Object = spcObject
	glog.V(4).Infof("Queuing SPC %s for add event", spcObject.Name)
	c.enqueueSpc(&c.queueLoad)
}

func (c *Controller) updateSpc(oldSpc, newSpc interface{}) {
	spcObjectNew := newSpc.(*apis.StoragePoolClaim)
	spcObjectOld := oldSpc.(*apis.StoragePoolClaim)

	if spcObjectNew.ObjectMeta.ResourceVersion == spcObjectOld.ObjectMeta.ResourceVersion {
		// If Resource Version is same it means the object has not got updated.
		c.queueLoad.Operation = ignoreEvent
	} else {
		// Suppressing delete event here as the event is already captured in
		// deleteSpc hook.
		if IsDeleteEvent(spcObjectNew) {
			c.queueLoad.Operation = ignoreEvent
		} else {
			// To-DO
			// Implement Logic for Update of SPC object
			c.queueLoad.Operation = updateEvent
			c.queueLoad.Object = spcObjectNew
			glog.V(4).Infof("Queuing SPC %s for update event", spcObjectNew.Name)
			c.enqueueSpc(&c.queueLoad)
		}

	}

}

func (c *Controller) deleteSpc(obj interface{}) {
	spcObject := obj.(*apis.StoragePoolClaim)
	c.queueLoad.Operation = deleteEvent
	c.queueLoad.Object = spcObject
	glog.V(4).Infof("Queuing SPC %s for delete event", spcObject.Name)
	c.enqueueSpc(&c.queueLoad)
}

// IsDeleteEvent is to check if the call is for SPC delete.
func IsDeleteEvent(spc *apis.StoragePoolClaim) bool {
	if spc.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

/*
Copyright 2019 The OpenEBS Authors

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

package cvc

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	listers "github.com/openebs/maya/pkg/client/generated/listers/openebs.io/v1alpha1"
	ndmclientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "cstorvolumeclaim-controller"

// Controller is the controller implementation for SPC resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// ndmclientset is a ndm custom resource package generated for custom API group.
	ndmclientset ndmclientset.Interface

	cvcLister listers.CStorVolumeClaimLister
	cvLister  listers.CStorVolumeLister
	cspLister listers.CStorPoolLister
	// cvcSynced is used for caches sync to get populated
	cvcSynced cache.InformerSynced

	// Store is a generic object storage interface. Reflector knows how to watch a server
	// and update a store. A generic store is provided, which allows Reflector to be used
	// as a local caching system, and an LRU store, which allows Reflector to work like a
	// queue of items yet to be processed.
	cvcStore cache.Store

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

// ControllerBuilder is the builder object for controller.
type ControllerBuilder struct {
	Controller *Controller
}

// NewControllerBuilder returns an empty instance of controller builder.
func NewControllerBuilder() *ControllerBuilder {
	return &ControllerBuilder{
		Controller: &Controller{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *ControllerBuilder) withKubeClient(ks kubernetes.Interface) *ControllerBuilder {
	cb.Controller.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *ControllerBuilder) withOpenEBSClient(cs clientset.Interface) *ControllerBuilder {
	cb.Controller.clientset = cs
	return cb
}

// withNDMClient fills ndm client to controller object.
func (cb *ControllerBuilder) withNDMClient(ndmcs ndmclientset.Interface) *ControllerBuilder {
	cb.Controller.ndmclientset = ndmcs
	return cb
}

// withCVCLister fills cvc lister to controller object.
func (cb *ControllerBuilder) withCVCLister(sl informers.SharedInformerFactory) *ControllerBuilder {
	cvcInformer := sl.Openebs().V1alpha1().CStorVolumeClaims()
	cb.Controller.cvcLister = cvcInformer.Lister()
	return cb
}

// withCVRLister fills cvr lister to controller object.
func (cb *ControllerBuilder) withCVLister(sl informers.SharedInformerFactory) *ControllerBuilder {
	cvInformer := sl.Openebs().V1alpha1().CStorVolumes()
	cb.Controller.cvLister = cvInformer.Lister()
	return cb
}

// withCSPLister fills csp lister to controller object.
func (cb *ControllerBuilder) withCSPLister(sl informers.SharedInformerFactory) *ControllerBuilder {
	cspInformer := sl.Openebs().V1alpha1().CStorPools()
	cb.Controller.cspLister = cspInformer.Lister()
	return cb
}

// withCVCLister returns a Store implemented simply with a map and a lock.
func (cb *ControllerBuilder) withCVCStore() *ControllerBuilder {
	cb.Controller.cvcStore = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	return cb
}

// withspcSynced adds object sync information in cache to controller object.
func (cb *ControllerBuilder) withCVCSynced(sl informers.SharedInformerFactory) *ControllerBuilder {
	cvcInformer := sl.Openebs().V1alpha1().CStorVolumeClaims()
	cb.Controller.cvcSynced = cvcInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *ControllerBuilder) withWorkqueueRateLimiting() *ControllerBuilder {
	cb.Controller.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CVC")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *ControllerBuilder) withRecorder(ks kubernetes.Interface) *ControllerBuilder {
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.Controller.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *ControllerBuilder) withEventHandler(spcInformerFactory informers.SharedInformerFactory) *ControllerBuilder {
	cvcInformer := spcInformerFactory.Openebs().V1alpha1().CStorVolumeClaims()
	// Set up an event handler for when CVC resources change
	cvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.Controller.addCVC,
		UpdateFunc: cb.Controller.updateCVC,
		DeleteFunc: cb.Controller.deleteCVC,
	})
	return cb
}

// Build returns a controller instance.
func (cb *ControllerBuilder) Build() (*Controller, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.Controller, nil
}

// addCVC is the add event handler for CstorVolumeClaim
func (c *Controller) addCVC(obj interface{}) {
	cvc, ok := obj.(*apis.CStorVolumeClaim)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get cvc object %#v", obj))
		return
	}

	glog.V(4).Infof("Queuing CVC %s for add event", cvc.Name)
	c.enqueueCVC(cvc)
}

// updateCVC is the update event handler for CstorVolumeClaim
func (c *Controller) updateCVC(oldCVC, newCVC interface{}) {
	_, ok := newCVC.(*apis.CStorVolumeClaim)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get cvc object %#v", newCVC))
		return
	}
	//if c.isCVCPending(cvc) {
	c.enqueueCVC(newCVC)
	//}
}

// deleteCVC is the delete event handler for CstorVolumeClaim
func (c *Controller) deleteCVC(obj interface{}) {
	cvc, ok := obj.(*apis.CStorVolumeClaim)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		cvc, ok = tombstone.Obj.(*apis.CStorVolumeClaim)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a cstorvolumeclaim %#v", obj))
			return
		}
	}
	glog.V(4).Infof("Deleting cstorvolumeclaim %s", cvc.Name)
	c.enqueueCVC(cvc)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CVC controller")

	// Wait for the k8s caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cvcSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	glog.Info("Starting CVC workers")
	// Launch worker to process CVC resources
	// Threadiness will decide the number of workers you want to launch to process work items from queue
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started CVC workers")
	<-stopCh
	glog.Info("Shutting down CVC workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.V(1).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

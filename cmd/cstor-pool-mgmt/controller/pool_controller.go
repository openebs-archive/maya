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

package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/volumereplica"
)

const poolControllerName = "CStorPool"

// CStorPoolController is the controller implementation for CStorPool resources.
type CStorPoolController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// cStorPoolSynced is used for caches sync to get populated
	cStorPoolSynced cache.InformerSynced

	// deletedIndexer holds deleted resource to be retrived after workqueue
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
}

// NewCStorPoolController returns a new instance of CStorPool controller
func NewCStorPoolController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorPoolController {

	// obtain references to shared index informers for the cStorPool resources
	cStorPoolInformer := cStorInformerFactory.Openebs().V1alpha1().CStorPools()

	openebsScheme.AddToScheme(scheme.Scheme)

	// Create event broadcaster to receive events and send them to any EventSink, watcher, or log.
	// Add NewCstorPoolController types to the default Kubernetes Scheme so Events can be
	// logged for CstorPool Controller types.
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: poolControllerName})

	controller := &CStorPoolController{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		deletedIndexer: cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc,
			cache.Indexers{}),
		cStorPoolSynced: cStorPoolInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorPool"),
		recorder:        recorder,
	}

	glog.Info("Setting up event handlers")

	// Instantiating QueueLoad before entering workqueue.
	q := QueueLoad{}

	// Set up an event handler for when CstorPool resources change.
	cStorPoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			q.operation = "add"
			controller.enqueueCStorPool(obj, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newCStorPool := new.(*apis.CStorPool)
			oldCStorPool := old.(*apis.CStorPool)
			// Periodic resync will send update events for all known CStorPool.
			// Two different versions of the same CStorPool will always have different RVs.
			if newCStorPool.ResourceVersion == oldCStorPool.ResourceVersion {
				return
			}
			q.operation = "update"
			controller.enqueueCStorPool(new, q)
		},
		DeleteFunc: func(obj interface{}) {
			q.operation = "delete"
			controller.enqueueCStorPool(obj, q)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *CStorPoolController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CStorPool controller")

	// Wait for the k8s caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cStorPoolSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	glog.Info("Starting CStorPool workers")
	// Launch worker to process CStorPool resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started CStorPool workers")
	<-stopCh
	glog.Info("Shutting down CStorPool workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *CStorPoolController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *CStorPoolController) processNextWorkItem() bool {
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
		var q QueueLoad
		var ok bool

		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if q, ok = obj.(QueueLoad); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// cStorPool resource to be synced.
		if err := c.syncHandler(q.key, q.operation); err != nil {
			return fmt.Errorf("error syncing '%s': %s", q.key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", q.key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorPoolUpdated resource
// with the current status of the resource.
func (c *CStorPoolController) syncHandler(key, operation string) error {
	cStorPoolUpdated, err := c.getPoolResource(key, operation)
	if err != nil {
		return err
	}
	switch operation {
	case "add":
		glog.Info("added event")

		err := pool.CheckValidPool(cStorPoolUpdated)
		if err != nil {
			return err
		}

		err = pool.ImportPool(cStorPoolUpdated)
		if err == nil {
			InitialImportedPoolVol = volumereplica.GetVolumes()
			return nil
		}

		err = pool.CreatePool(cStorPoolUpdated)
		if err != nil {
			return err
		}
		break

	case "update":
		glog.Info("updated event")
		break

	case "delete":
		glog.Info("deleted event")
		err := pool.DeletePool(cStorPoolUpdated.Spec.PoolSpec.PoolName)
		if err != nil {
			return err
		}
		break
	}

	return nil
}

// enqueueCstorPool takes a CstorPool resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CstorPools.
func (c *CStorPoolController) enqueueCStorPool(obj interface{}, q QueueLoad) {
	var key string
	var err error

	if q.operation == "delete" {
		c.deletedIndexer.Add(obj)
		key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
		if err != nil {
			runtime.HandleError(err)
			return
		}
	} else {
		if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
			runtime.HandleError(err)
			return
		}
	}
	q.key = key
	c.workqueue.AddRateLimited(q)
}

// getPoolResource returns object corresponding to the resource key
func (c *CStorPoolController) getPoolResource(key, operation string) (*apis.CStorPool, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	if operation == "delete" {
		if obj, exists, err := c.deletedIndexer.GetByKey(key); err == nil && exists {
			c.deletedIndexer.Delete(key)
			return obj.(*apis.CStorPool), nil
		}
	}
	cStorPoolUpdated, err := c.clientset.OpenebsV1alpha1().CStorPools().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorPoolUpdated '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return cStorPoolUpdated, nil
}

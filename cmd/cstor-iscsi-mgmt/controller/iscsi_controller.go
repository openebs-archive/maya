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

	"github.com/openebs/maya/cmd/cstor-iscsi-mgmt/cstorops/iscsi"
)

const iscsiControllerName = "CStorIscsi"

// CStorIscsiController is the controller implementation for CStorIscsi resources.
type CStorIscsiController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// cStorIscsiSynced is used for caches sync to get populated
	cStorIscsiSynced cache.InformerSynced

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

// NewCStorIscsiController returns a new instance of CStorIscsi controller
func NewCStorIscsiController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorIscsiController {

	// obtain references to shared index informers for the cStorIscsi resources
	cStorIscsiInformer := cStorInformerFactory.Openebs().V1alpha1().CStorVolumes()

	openebsScheme.AddToScheme(scheme.Scheme)

	// Create event broadcaster to receive events and send them to any EventSink, watcher, or log.
	// Add NewCstorIscsiController types to the default Kubernetes Scheme so Events can be
	// logged for CstorIscsi Controller types.
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: iscsiControllerName})

	controller := &CStorIscsiController{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		deletedIndexer: cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc,
			cache.Indexers{}),
		cStorIscsiSynced: cStorIscsiInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorIscsi"),
		recorder:         recorder,
	}

	glog.Info("Setting up event handlers")

	// Instantiating QueueLoad before entering workqueue.
	q := QueueLoad{}

	// Set up an event handler for when CstorIscsi resources change.
	cStorIscsiInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			q.operation = "add"
			controller.enqueueCStorIscsi(obj, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newCStorIscsi := new.(*apis.CStorVolume)
			oldCStorIscsi := old.(*apis.CStorVolume)
			// Periodic resync will send update events for all known CStorIscsi.
			// Two different versions of the same CStorIscsi will always have different RVs.
			if newCStorIscsi.ResourceVersion == oldCStorIscsi.ResourceVersion {
				return
			}
			q.operation = "update"
			controller.enqueueCStorIscsi(new, q)
		},
		DeleteFunc: func(obj interface{}) {
			q.operation = "delete"
			controller.enqueueCStorIscsi(obj, q)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *CStorIscsiController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CStorIscsi controller")

	// Wait for the k8s caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cStorIscsiSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	glog.Info("Starting CStorIscsi workers")
	// Launch worker to process CStorIscsi resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started CStorIscsi workers")
	<-stopCh
	glog.Info("Shutting down CStorIscsi workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *CStorIscsiController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *CStorIscsiController) processNextWorkItem() bool {
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
		// cStorIscsi resource to be synced.
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
// converge the two. It then updates the Status block of the cStorIscsiUpdated resource
// with the current status of the resource.
func (c *CStorIscsiController) syncHandler(key, operation string) error {
	cStorIscsiUpdated, err := c.getIscsiResource(key, operation)
	if err != nil {
		return err
	}
	switch operation {
	case "add":
		glog.Info("added event")

		err := iscsi.CheckValidIscsi(cStorIscsiUpdated)
		if err != nil {
			return err
		}

		err = iscsi.CreateIscsi(cStorIscsiUpdated)
		if err != nil {
			return err
		}
		break

	case "update":
		glog.Info("updated event")
		err := iscsi.CheckValidIscsi(cStorIscsiUpdated)
		if err != nil {
			return err
		}

		err = iscsi.CreateIscsi(cStorIscsiUpdated)
		if err != nil {
			return err
		}
		break

	case "delete":
		glog.Info("deleted event")
		break

	}

	return nil
}

// enqueueCstorIscsi takes a CstorIscsi resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CstorIscsis.
func (c *CStorIscsiController) enqueueCStorIscsi(obj interface{}, q QueueLoad) {
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

// getIscsiResource returns object corresponding to the resource key
func (c *CStorIscsiController) getIscsiResource(key, operation string) (*apis.CStorVolume, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	if operation == "delete" {
		if obj, exists, err := c.deletedIndexer.GetByKey(key); err == nil && exists {
			c.deletedIndexer.Delete(key)
			return obj.(*apis.CStorVolume), nil
		}
	}
	cStorIscsiUpdated, err := c.clientset.OpenebsV1alpha1().CStorVolumes().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorIscsi resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorIscsiUpdated '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return cStorIscsiUpdated, nil
}

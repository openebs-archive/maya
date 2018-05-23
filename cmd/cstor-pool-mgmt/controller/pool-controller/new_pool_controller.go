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

package poolcontroller

import (
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
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
	q := common.QueueLoad{}

	// Set up an event handler for when CstorPool resources change.
	cStorPoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if !IsRightCStorPoolMgmt(obj.(*apis.CStorPool)) {
				return
			}
			q.Operation = "add"
			glog.Infof("added event")
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
			if !IsRightCStorPoolMgmt(newCStorPool) {
				return
			}

			if IsOnlyStatusChange(oldCStorPool, newCStorPool) {
				return
			}
			if IsDestroyEvent(newCStorPool) {
				q.Operation = "destroy"
			} else {
				q.Operation = "modify"
				return // will be removed once modify is implemented
			}
			controller.enqueueCStorPool(new, q)
		},
		DeleteFunc: func(obj interface{}) {
			glog.Infof("\nk8s-deleted event")
		},
	})

	return controller
}

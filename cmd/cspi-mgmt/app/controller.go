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

package app

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	zpool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	//openebsScheme "github.com/openebs/maya/pkg/client/generated/openebs.io/v1alpha2/clientset/internalclientset/scheme"
)

const poolControllerName = "CStorPoolInstance"

// CStorPoolInstanceController is the controller implementation for CStorPoolInstance resources.
type CStorPoolInstanceController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// cStorPoolInstanceSynced is used for caches sync to get populated
	cStorPoolInstanceSynced cache.InformerSynced

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

// NewCStorPoolInstanceController returns a new instance of CStorPoolInstance controller
func NewCStorPoolInstanceController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorPoolInstanceController {

	// obtain references to shared index informers for the cStorPoolInstance resources
	cStorPoolInstanceInformer := cStorInformerFactory.Openebs().V1alpha1().CStorPoolInstances()

	zpool.KubeClient = kubeclientset
	zpool.OpenEBSClient = clientset

	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Errorf("failed to add to scheme: error {%v}", err)
		return nil
	}

	// Create event broadcaster to receive events and send them to any EventSink, watcher, or log.
	// Add NewCstorPoolInstanceController types to the default Kubernetes Scheme so Events can be
	// logged for CstorPoolInstance Controller types.
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: poolControllerName})

	controller := &CStorPoolInstanceController{
		kubeclientset:           kubeclientset,
		clientset:               clientset,
		cStorPoolInstanceSynced: cStorPoolInstanceInformer.Informer().HasSynced,
		workqueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), poolControllerName),
		recorder:                recorder,
	}

	klog.Info("Setting up event handlers for CSP")

	// Set up an event handler for when CstorPoolInstance resources change.
	cStorPoolInstanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cspi := obj.(*apis.CStorPoolInstance)
			if !IsRightCStorPoolInstanceMgmt(cspi) {
				return
			}
			controller.enqueueCStorPoolInstance(cspi)
		},

		UpdateFunc: func(oldVar, newVar interface{}) {
			cspi := newVar.(*apis.CStorPoolInstance)

			if !IsRightCStorPoolInstanceMgmt(cspi) {
				return
			}
			controller.enqueueCStorPoolInstance(cspi)
		},
	})

	return controller
}

// enqueueCstorPoolInstance takes a CStorPoolInstance resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorPoolInstances.
func (c *CStorPoolInstanceController) enqueueCStorPoolInstance(obj *apis.CStorPoolInstance) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(common.QueueLoad{Key: key})
}

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
	"fmt"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	listers "github.com/openebs/maya/pkg/client/generated/listers/openebs.io/v1alpha1"
	ndmclientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "spc-controller"

// Controller is the controller implementation for SPC resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// ndmclientset is a ndm custom resource package generated for custom API group.
	ndmclientset ndmclientset.Interface

	spcLister listers.StoragePoolClaimLister

	// spcSynced is used for caches sync to get populated
	spcSynced cache.InformerSynced

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

// withSpcLister fills spc lister to controller object.
func (cb *ControllerBuilder) withSpcLister(sl informers.SharedInformerFactory) *ControllerBuilder {
	spcInformer := sl.Openebs().V1alpha1().StoragePoolClaims()
	cb.Controller.spcLister = spcInformer.Lister()
	return cb
}

// withspcSynced adds object sync information in cache to controller object.
func (cb *ControllerBuilder) withspcSynced(sl informers.SharedInformerFactory) *ControllerBuilder {
	spcInformer := sl.Openebs().V1alpha1().StoragePoolClaims()
	cb.Controller.spcSynced = spcInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *ControllerBuilder) withWorkqueueRateLimiting() *ControllerBuilder {
	cb.Controller.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SPC")
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
	spcInformer := spcInformerFactory.Openebs().V1alpha1().StoragePoolClaims()
	// Set up an event handler for when SPC resources change
	spcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.Controller.addSpc,
		UpdateFunc: cb.Controller.updateSpc,
		// This will enter the sync loop and no-op, because the spc has been deleted from the store.
		DeleteFunc: cb.Controller.deleteSpc,
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

// addSpc is the add event handler for spc
func (c *Controller) addSpc(obj interface{}) {
	spc, ok := obj.(*apis.StoragePoolClaim)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get spc object %#v", obj))
		return
	}
	if spc.Annotations[string(apis.OpenEBSDisableReconcileKey)] == "true" {
		message := fmt.Sprintf("reconcile is disabled via %q annotation", string(apis.OpenEBSDisableReconcileKey))
		c.recorder.Event(spc, corev1.EventTypeWarning, "Create", message)
		return
	}
	glog.V(4).Infof("Queuing SPC %s for add event", spc.Name)
	c.enqueueSpc(spc)
}

// updateSpc is the update event handler for spc.
func (c *Controller) updateSpc(oldSpc, newSpc interface{}) {
	spc, ok := newSpc.(*apis.StoragePoolClaim)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get spc object %#v", newSpc))
		return
	}
	if spc.Annotations[string(apis.OpenEBSDisableReconcileKey)] == "true" {
		message := fmt.Sprintf("reconcile is disabled via %q annotation", string(apis.OpenEBSDisableReconcileKey))
		c.recorder.Event(spc, corev1.EventTypeWarning, "Update", message)
		return
	}
	// Enqueue spc only when there is a pending pool to be created.
	if c.isPoolPending(spc) {
		c.enqueueSpc(newSpc)
	}
}

// deleteSpc is the delete event handler for spc.
func (c *Controller) deleteSpc(obj interface{}) {
	spc, ok := obj.(*apis.StoragePoolClaim)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		spc, ok = tombstone.Obj.(*apis.StoragePoolClaim)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a storagepoolclaim %#v", obj))
			return
		}
	}
	if spc.Annotations[string(apis.OpenEBSDisableReconcileKey)] == "true" {
		message := fmt.Sprintf("reconcile is disabled via %q annotation", string(apis.OpenEBSDisableReconcileKey))
		c.recorder.Event(spc, corev1.EventTypeWarning, "Delete", message)
		return
	}
	glog.V(4).Infof("Deleting storagepoolclaim %s", spc.Name)
	c.enqueueSpc(spc)
}

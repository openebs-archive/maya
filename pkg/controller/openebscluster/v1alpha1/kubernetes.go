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

package v1alpha1

import (
	"context"

	apisoecluster "github.com/openebs/maya/pkg/apis/openebs.io/openebscluster/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	oecluster "github.com/openebs/maya/pkg/openebscluster/v1alpha1"

	log "github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// ControllerName is the name given to this controller
	//
	// NOTE:
	//  This name can be used by resources if they
	// want to be managed by this controller
	ControllerName string = "openebscluster-controller"

	// SelfNamespace is the environment variable pointing
	// at the namespace under which this controller
	// will be running
	SelfNamespace string = "OPENEBS_IO_SELF_NAMESPACE"
)

// kubeReconciler is kubernetes based openebscluster
// reconciler
type kubeReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a
// OpenebsCluster object and makes changes based on
// the cluster state and what is in the
// OpenebsCluster's Spec
func (r *kubeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// first task is to fetch requested openebs cluster
	// instance
	instance := &apisoecluster.OpenebsCluster{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// object is not found at this point in time
			// nothing to do
			return reconcile.Result{}, nil
		}
		// error reading the object - requeue the request
		return reconcile.Result{}, err
	}
	// TODO
	// business logic to be written
	return reconcile.Result{}, nil
}

// KubeRegister registers this controller against
// the provided kubernetes based controller manager
// instance
func KubeRegister(mgr manager.Manager) error {
	return KubeController().register(mgr)
}

type kubeController struct{}

// KubeController returns a new instance of
// kubernetes based openebscluster controller
func KubeController() *kubeController {
	return &kubeController{}
}

// register registers the provided kubernetes
// manager against this controller
//
// NOTE:
//  All the required watches need to be set here
func (k *kubeController) register(mgr manager.Manager) error {
	r := &kubeReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// namespace under which this controller is running
	//
	// NOTE:
	//  This helps this controller to watch the resources
	// tagged with managed by namespace by comparing
	// the tagged namespace with its own namespace. If
	// match is a success then that resource will be
	// watched else ignored.
	ns := env.Get(env.ENVKey(SelfNamespace))
	if ns == "" {
		log.Errorf("may skip watching resources: environment variable '%s' not set", SelfNamespace)
	}
	// watch for changes to OpenebsCluster resource(s)
	//
	// NOTE:
	//  Predicate ensures specific OpenebsCluster resources are
	// only enqueued
	//
	// NOTE:
	//  this controller is not an owner of OpenebsCluster
	// resource. However, OpenebsCluster resource can be set
	// to be managed by this controller via later's name &/or
	// namespace
	return c.Watch(
		&source.Kind{Type: &apisoecluster.OpenebsCluster{}},
		&handler.EnqueueRequestForObject{},
		oecluster.Predicate(
			oecluster.IsControlledByNameIfSet(ControllerName),
			oecluster.IsControlledByNamespaceIfSet(ns),
		),
	)
}

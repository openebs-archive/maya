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

package cstorvolumeclaim

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"time"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	ndmclientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	leader "github.com/openebs/maya/pkg/kubernetes/leaderelection"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
	// lease lock resource name for lease API resource
	leaderElectionLockName = "cvc-controller-leader"
)

// Command line flags
var (
	leaderElection          = flag.Bool("leader-election", false, "Enables leader election.")
	leaderElectionNamespace = flag.String("leader-election-namespace", "", "The namespace where the leader election resource exists. Defaults to the pod namespace if not set.")
)

// Start starts the cstorvolumeclaim controller.
func Start() error {

	klog.InitFlags(nil)
	err := flag.Set("logtostderr", "true")
	if err != nil {
		return errors.Wrap(err, "failed to set logtostderr flag")
	}
	flag.Parse()

	// Get in cluster config
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "error building kubeconfig")
	}

	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building kubernetes clientset")
	}

	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building openebs clientset")
	}

	// Building NDM Clientset
	ndmClient, err := ndmclientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building ndm clientset")
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	cvcInformerFactory := informers.NewSharedInformerFactory(openebsClient, time.Second*30)

	// Build() fn of all controllers calls AddToScheme to adds all types of this
	// clientset into the given scheme.
	// If multiple controllers happen to call this AddToScheme same time,
	// it causes panic with error saying concurrent map access.
	// This lock is used to serialize the AddToScheme call of all controllers.
	controller, err := NewCVCControllerBuilder().
		withKubeClient(kubeClient).
		withOpenEBSClient(openebsClient).
		withNDMClient(ndmClient).
		withCVCSynced(cvcInformerFactory).
		withCVCLister(cvcInformerFactory).
		withCVLister(cvcInformerFactory).
		withCVRLister(cvcInformerFactory).
		withCVRInformerSync(cvcInformerFactory).
		withCVCStore().
		withRecorder(kubeClient).
		withEventHandler(cvcInformerFactory).
		withWorkqueueRateLimiting().Build()

	if err != nil {
		return errors.Wrapf(err, "error building controller instance")
	}

	// Threadiness defines the number of workers to be launched in Run function
	run := func(context.Context) {
		// run...
		stopCh := make(chan struct{})
		kubeInformerFactory.Start(stopCh)
		cvcInformerFactory.Start(stopCh)
		go controller.Run(2, stopCh)

		// ...until SIGINT
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		close(stopCh)
	}

	if !*leaderElection {
		run(context.TODO())
	} else {
		le := leader.NewLeaderElection(kubeClient, leaderElectionLockName, run)
		if *leaderElectionNamespace != "" {
			le.WithNamespace(*leaderElectionNamespace)
		}
		if err := le.Run(); err != nil {
			klog.Fatalf("failed to initialize leader election: %v", err)
		}
	}
	return nil
}

// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to get k8s Incluster config. %+v", err)
		if kubeconfig == "" {
			return nil, errors.Wrap(err, "kubeconfig is empty")
		}
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			return nil, errors.Wrap(err, "error building kubeconfig")
		}
	}
	return cfg, err
}

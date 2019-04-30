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
	"github.com/pkg/errors"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	masterURL  string
	kubeconfig string
)

// Start starts the cstor-operator.
func Start() error {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

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

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	spcInformerFactory := informers.NewSharedInformerFactory(openebsClient, time.Second*30)
	controller, err := NewControllerBuilder().
		withKubeClient(kubeClient).
		withOpenEBSClient(openebsClient).
		withspcSynced(spcInformerFactory).
		withSpcLister(spcInformerFactory).
		withRecorder(kubeClient).
		withEventHandler(spcInformerFactory).
		withWorkqueueRateLimiting().Build()
	if err != nil {
		return errors.Wrapf(err, "error building controller instance")
	}

	go kubeInformerFactory.Start(stopCh)
	go spcInformerFactory.Start(stopCh)

	// Threadiness defines the number of workers to be launched in Run function
	return controller.Run(2, stopCh)
}

// Cannot be unit tested
// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	var masterURL string
	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Failed to get k8s Incluster config. %+v", err)
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

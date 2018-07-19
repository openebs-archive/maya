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

package spc

import (
	"time"
	"github.com/golang/glog"
	"k8s.io/client-go/rest"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	"fmt"
)

var (
	masterURL  string
	kubeconfig string
)
type QueueLoad struct {
	Key       string
	Operation string
}

func Start() (error) {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// Get in cluster config
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %s", err.Error())
	}

	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building openebs clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	spcInformerFactory := informers.NewSharedInformerFactory(openebsClient, time.Second*30)

	controller := NewController(kubeClient, openebsClient, kubeInformerFactory, spcInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go spcInformerFactory.Start(stopCh)

	// Threadiness defines the nubmer of workers to be launched in Run function
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
			return nil, fmt.Errorf("kubeconfig is empty: %v", err.Error())
		}
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
		}
	}
	return cfg, err
}
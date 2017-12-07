/*
Copyright 2017 The Kubernetes Authors.

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

package mcrd

import (
	"time"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/openebs/maya/pkg/signals"
	spclientset "github.com/openebs/maya/pkg/storagepool-client/clientset/versioned"
	spcclientset "github.com/openebs/maya/pkg/storagepoolclaim-client/clientset/versioned"
	spcinformers "github.com/openebs/maya/pkg/storagepoolclaim-client/informers/externalversions"
)

func Checkcrd(kuberconfig string) {
	masterURL := ""
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kuberconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	spcClient, err := spcclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building storage pool claim clientset: %s", err.Error())
	}

	spClient, err := spclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building  storage pool Client: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := spcinformers.NewSharedInformerFactory(spcClient, time.Second*30)

	controller := NewController(kubeClient, spcClient, spClient, kubeInformerFactory,
		exampleInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go exampleInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

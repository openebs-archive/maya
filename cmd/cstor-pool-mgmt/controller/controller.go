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
	"sync"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/pool"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
)

// StartControllers instantiates CStorPool and CStorVolumeReplica controllers
// and watches them.
func StartControllers(kubeconfig string) {
	// Set up signals to handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		glog.Fatalf(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building openebs clientset: %s", err.Error())
	}

	// Blocking call for checking status of zrepl running in cstor-pool container.
	pool.CheckForZrepl()

	// Blocking call for checking status of CStorPool CRD.
	checkForCStorPoolCRD(openebsClient)

	// Blocking call for checking status of CStorVolumeReplica CRD.
	checkForCStorVolumeReplicaCRD(openebsClient)

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(openebsClient, time.Second*30)

	// Instantiate the cStor Pool and VolumeReplica controllers.
	poolController := NewCStorPoolController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	volumeReplicaController := NewCStorVolumeReplicaController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go openebsInformerFactory.Start(stopCh)

	// Waitgroup for starting pool and VolumeReplica controller goroutines.
	var wg sync.WaitGroup
	wg.Add(2)

	// Run controller for cStorPool.
	go func() {
		if err = poolController.Run(2, stopCh); err != nil {
			glog.Fatalf("Error running CStorPool controller: %s", err.Error())
		}
		wg.Done()
	}()

	// PoolNameHandler tries to get pool name and blocks for
	// particular number of attempts.
	var noOfAttempts int = 3
	PoolNameHandler(noOfAttempts)

	// Run controller for cStorVolumeReplica.
	go func() {
		if err = volumeReplicaController.Run(2, stopCh); err != nil {
			glog.Fatalf("Error running CStorVolumeReplica controller: %s", err.Error())
		}
		wg.Done()
	}()
	wg.Wait()

}

// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	var masterURL string
	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("failed to get k8s Incluster config. %+v", err)
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

// checkForCStorPoolCRD is Blocking call for checking status of CStorPool CRD.
func checkForCStorPoolCRD(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorPools CRD not found...")
			time.Sleep(10 * time.Second)
			continue
		}
		glog.Info("CStorPool CRD found")
		break
	}
}

// checkForCStorVolumeReplicaCRD is Blocking call for checking status of CStorVolumeReplica CRD.
func checkForCStorVolumeReplicaCRD(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorVolumeReplicas().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorVolumeReplicas CRD not found...")
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
}

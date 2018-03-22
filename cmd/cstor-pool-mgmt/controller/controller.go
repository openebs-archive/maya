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

package crdops

import (
	"os/exec"
	"sync"
	"time"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
)

// StartControllers instantiates CRD controllers and watches them.
func StartControllers(kuberconfig string) {
	masterURL := ""
	// Set up signals to handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("failed to get k8s Incluster config. %+v", err)
		if kuberconfig == "" {
			glog.Fatalf("kubeconfig is empty")
		} else {
			cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kuberconfig)
			if err != nil {
				glog.Fatalf("Error building kubeconfig: %s", err.Error())
			}
		}
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	crdClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building sp-spc clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(crdClient, time.Second*30)

	// Instantiate the cStor Pool and VolumeReplica controllers.
	poolController := NewCStorPoolController(kubeClient, crdClient, kubeInformerFactory,
		exampleInformerFactory)

	volumeReplicaController := NewCStorVolumeReplicaController(kubeClient, crdClient, kubeInformerFactory,
		exampleInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go exampleInformerFactory.Start(stopCh)

	// Blocking call for checking status of zrepl running in cstor-pool container.
	CheckForZrepl()

	// Waitgroup for starting pool and VolumeReplica controller goroutines.
	var wg sync.WaitGroup
	wg.Add(2)

	// Run controller for cStorPool.
	go func() {
		if err = poolController.Run(2, stopCh); err != nil {
			glog.Fatalf("Error running cstor controller: %s", err.Error())
		}
		wg.Done()
	}()

	time.Sleep(2 * time.Second)

	// Run controller for cStorReplica.
	go func() {
		if err = volumeReplicaController.Run(2, stopCh); err != nil {
			glog.Fatalf("Error running cstor controller: %s", err.Error())
		}
		wg.Done()
	}()
	wg.Wait()

}

// CheckForZrepl is blocking call for checking status of zrepl in cstor-main container.
func CheckForZrepl() {
	for {
		statuscmd := exec.Command("zpool", "status")
		_, err := statuscmd.CombinedOutput()
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for zrepl...")
			continue
		}
		break
	}
}

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

package startcontroller

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	poolcontroller "github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/pool-controller"
	replicacontroller "github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/replica-controller"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	"github.com/openebs/maya/pkg/util"
)

const (
	// NumThreads defines number of worker threads for resource watcher.
	NumThreads = 2
	// NumRoutinesThatFollow is for handling golang waitgroups.
	NumRoutinesThatFollow = 1
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

	// Making RunnerVar to use RealRunner
	pool.RunnerVar = util.RealRunner{}
	volumereplica.RunnerVar = util.RealRunner{}
	common.IsImported = make(chan bool, 1)

	// Blocking call for checking status of zrepl running in cstor-pool container.
	pool.CheckForZrepl()

	// Blocking call for checking status of CStorPool CRD.
	common.CheckForCStorPoolCRD(openebsClient)

	// Blocking call for checking status of CStorVolumeReplica CRD.
	common.CheckForCStorVolumeReplicaCRD(openebsClient)

	// NewSharedInformerFactory constructs a new instance of k8s sharedInformerFactory.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, common.SharedInformerInterval)
	// openebsInformerFactory constructs a new instance of openebs sharedInformerFactory.
	openebsInformerFactory := informers.NewSharedInformerFactory(openebsClient, common.SharedInformerInterval)

	// Instantiate the cStor Pool and VolumeReplica controllers.
	cStorPoolController := poolcontroller.NewCStorPoolController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	volumeReplicaController := replicacontroller.NewCStorVolumeReplicaController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go openebsInformerFactory.Start(stopCh)

	// Waitgroup for starting pool and VolumeReplica controller goroutines.
	var wg sync.WaitGroup
	wg.Add(NumRoutinesThatFollow)

	// Run controller for cStorPool.
	go func() {
		if err = cStorPoolController.Run(NumThreads, stopCh); err != nil {
			glog.Fatalf("Error running CStorPool controller: %s", err.Error())
		}
		wg.Done()
	}()

	// CheckForCStorPool tries to get pool name and blocks forever because
	// volumereplica can be created only if pool is present.
	common.CheckForCStorPool()

	wg.Add(NumRoutinesThatFollow)
	// Run controller for cStorVolumeReplica.
	go func() {
		if err = volumeReplicaController.Run(NumThreads, stopCh); err != nil {
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
		glog.Errorf("Failed to get k8s Incluster config. %+v", err)
		if len(kubeconfig) == 0 {
			return nil, fmt.Errorf("kubeconfig is empty: %v", err.Error())
		}
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
		}
	}
	return cfg, err
}

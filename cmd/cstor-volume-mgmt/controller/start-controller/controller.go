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
	"os"
	"strconv"
	"sync"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	volumecontroller "github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/volume-controller"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	"github.com/openebs/maya/pkg/util"
)

const (
	// NumThreads defines number of worker threads for resource watcher.
	NumThreads = 1
	// NumRoutinesThatFollow is for handling golang waitgroups.
	NumRoutinesThatFollow = 1
)

// StartControllers instantiates CStorVolume controllers
// and watches them.
func StartControllers(kubeconfig string) {
	// Set up signals to handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		klog.Fatalf(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building openebs clientset: %s", err.Error())
	}

	volume.FileOperatorVar = util.RealFileOperator{}

	volume.UnixSockVar = util.RealUnixSock{}

	// Blocking call for checking status of istgt running in cstor-volume container.
	util.CheckForIscsi(volume.UnixSockVar)

	// Blocking call for checking status of CStorVolume CR.
	common.CheckForCStorVolumeCRD(openebsClient)

	// NewInformer returns a cache.Store and a controller for populating the store
	// while also providing event notifications. It’s basically a controller with some
	// boilerplate code to sync events from the FIFO queue to the downstream store.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, getSyncInterval())
	openebsInformerFactory := informers.NewSharedInformerFactory(openebsClient, getSyncInterval())

	cStorVolumeController := volumecontroller.NewCStorVolumeController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go openebsInformerFactory.Start(stopCh)

	// Waitgroup for starting volume controller goroutines.
	var wg sync.WaitGroup
	wg.Add(NumRoutinesThatFollow)

	// Run controller for cStorVolume.
	go func() {
		if err = cStorVolumeController.Run(NumThreads, stopCh); err != nil {
			klog.Fatalf("Error running CStorVolume controller: %s", err.Error())
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
		klog.Errorf("Failed to get k8s Incluster config. %+v", err)
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

// getSyncInterval gets the resync interval from environment variable.
// If missing or zero then default to DefaultSharedInformerInterval
// otherwise return the obtained value
func getSyncInterval() time.Duration {
	resyncInterval, err := strconv.Atoi(os.Getenv("RESYNC_INTERVAL"))
	if err != nil || resyncInterval == 0 {
		klog.Warningf("Incorrect resync interval %q obtained from env, defaulting to %q seconds", resyncInterval, common.DefaultSharedInformerInterval)
		return common.DefaultSharedInformerInterval
	}
	return time.Duration(resyncInterval) * time.Second
}

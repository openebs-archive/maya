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

package app

import (
	"flag"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
	backupcontroller "github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/backup-controller"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	replicacontroller "github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/replica-controller"
	restorecontroller "github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/restore"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	"github.com/pkg/errors"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "Path for kube config")
)

const (
	// NumThreads defines number of worker threads for resource watcher.
	NumThreads = 1
	// NumRoutinesThatFollow is for handling golang waitgroups.
	NumRoutinesThatFollow = 1
)

// Start starts the cspi-mgmt controller.
func Start() error {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	err := flag.Set("logtostderr", "true")
	if err != nil {
		return errors.Wrap(err, "failed to set logtostderr flag")
	}
	flag.Parse()

	cfg, err := getClusterConfig(*kubeconfig)
	if err != nil {
		return errors.Wrap(err, "error building kubeconfig")
	}

	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building kubernetes clientset")
	}
	common.Init()
	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building openebs clientset")
	}
	pool.CheckForZreplInitial(common.InitialZreplRetryInterval)
	/*
	   * go func() {
	   		// CheckForZreplContinuous is continuous health checker for status of
	   		// zrepl in cstor-pool container.
	   		// When zrepl is getting terminated and restarted very fast: zpool status
	   		// goroutine may miss this failure. To resolve, we’ll give InitialTimeDelay y
	   		// for zrepl container such that the period(x) of the goroutine thread will
	   		// be half that of this initialTimeDelay y. (x = 1/2 y).
	   		pool.CheckForZreplContinuous(common.ContinuousZreplRetryInterval)
	   		glog.Errorf("Zrepl/Pool is not available, Shutting down")
	   		os.Exit(1)
	   	}()
	*/

	// NewSharedInformerFactory constructs a new instance of k8s sharedInformerFactory.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, getSyncInterval())

	// openebsInformerFactory constructs a new instance of openebs sharedInformerFactory.
	openebsInformerFactory := informers.NewSharedInformerFactory(openebsClient, getSyncInterval())

	// Instantiate the cStor Pool Instance and VolumeReplica controllers.
	cStorPoolInstanceController := NewCStorPoolInstanceController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)
	volumeReplicaController := replicacontroller.NewCStorVolumeReplicaController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	// Instantiate the cStor backup controller
	backupController := backupcontroller.NewCStorBackupController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	// Instantiate the cStor restore controller
	restoreController := restorecontroller.NewCStorRestoreController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go openebsInformerFactory.Start(stopCh)
	// Blocking call for checking status of zrepl running in cstor-pool container.

	// Waitgroup for starting pool and VolumeReplica controller goroutines.
	var wg sync.WaitGroup
	//TODO: Remove below code

	wg.Add(NumRoutinesThatFollow)

	// Run controller for cStorPoolInstance
	go func() {
		if err = cStorPoolInstanceController.Run(NumThreads, stopCh); err != nil {
			glog.Fatalf("Error running CStorPoolInstance controller: %s", err.Error())
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

	wg.Add(NumRoutinesThatFollow)
	// Run controller for CStorBackup
	go func() {
		if err = backupController.Run(NumThreads, stopCh); err != nil {
			glog.Fatalf("Error running CStorBackup controller: %s", err.Error())
		}
		wg.Done()
	}()

	wg.Add(NumRoutinesThatFollow)
	// Run controller for CStorRestore.
	go func() {
		if err = restoreController.Run(NumThreads, stopCh); err != nil {
			glog.Fatalf("Error running CStorRestore controller: %s", err.Error())
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	glog.V(2).Info("Kubeconfig flag is empty")
	return rest.InClusterConfig()
}

// getSyncInterval gets the resync interval from environment variable.
// If missing or zero then default to SharedInformerInterval
// otherwise return the obtained value
func getSyncInterval() time.Duration {
	resyncInterval, err := strconv.Atoi(os.Getenv("RESYNC_INTERVAL"))
	if err != nil || resyncInterval == 0 {
		glog.Warningf("Incorrect resync interval %q obtained from env, defaulting to %q seconds", resyncInterval, common.SharedInformerInterval)
		return common.SharedInformerInterval
	}
	return time.Duration(resyncInterval) * time.Second
}

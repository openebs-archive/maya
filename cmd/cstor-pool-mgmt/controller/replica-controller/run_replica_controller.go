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

package replicacontroller

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *CStorVolumeReplicaController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CStorVolumeReplica controller")

	// Wait for the k8s caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cStorReplicaSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting CStorVolumeReplica workers")

	// Launch two workers to process CStorReplica resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, common.ResourceWorkerInterval, stopCh)
	}

	glog.Info("Started CStorVolumeReplica workers")
	<-stopCh
	glog.Info("Shutting down CStorVolumeReplica workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *CStorVolumeReplicaController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *CStorVolumeReplicaController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		var q common.QueueLoad
		var ok bool

		// We expect a particular object type to come off the workqueue.
		// These are of the form namespace/name. We do this as the delayed
		// nature of the workqueue means the items in the informer cache
		// may actually be more up to date that when the item was initially
		// put onto the workqueue.
		if q, ok = obj.(common.QueueLoad); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(
				errors.Errorf(
					"failed to process workqueue: got unsupported obj {%#v}",
					obj,
				),
			)
			return nil
		}

		// run syncHandler, passing it the namespace/name string
		// of cvr resource to be synced
		if err := c.syncHandler(q.Key, q.Operation); err != nil {
			return errors.Wrapf(
				err,
				"failed to process workqueue item {%s} during {%s} operation",
				q.Key,
				string(q.Operation),
			)
		}

		// Finally, if no error occurs we forget this item so that it
		// does not get queued again until another change happens
		c.workqueue.Forget(obj)
		glog.Infof(
			"workqueue item {%s} processed successfully for {%s} operation",
			q.Key,
			string(q.Operation),
		)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

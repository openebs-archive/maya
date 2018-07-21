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

package volumecontroller

import (
	"fmt"
	"os"
	"reflect"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorVolumeUpdated resource
// with the current status of the resource.
func (c *CStorVolumeController) syncHandler(key string, operation common.QueueOperation) error {
	glog.Infof("Handling %v operation for resource : %s ", operation, key)
	cStorVolumeGot, err := c.getVolumeResource(key)
	if err != nil {
		return err
	}
	status, err := c.cStorVolumeEventHandler(operation, cStorVolumeGot)
	if status == common.CVStatusIgnore {
		return nil
	}
	cStorVolumeGot.Status.Phase = apis.CStorVolumePhase(status)
	if err != nil {
		glog.Errorf(err.Error())
		glog.Infof("cStorVolume:%v, %v; Status: %v", cStorVolumeGot.Name,
			string(cStorVolumeGot.GetUID()), cStorVolumeGot.Status.Phase)

		_, err := c.clientset.OpenebsV1alpha1().CStorVolumes(cStorVolumeGot.Namespace).Update(cStorVolumeGot)
		if err != nil {
			return err
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorVolumes(cStorVolumeGot.Namespace).Update(cStorVolumeGot)
	if err != nil {
		return err
	}
	glog.Infof("cStorVolume:%v, %v; Status: %v", cStorVolumeGot.Name,
		string(cStorVolumeGot.GetUID()), cStorVolumeGot.Status.Phase)
	return nil

}

// cStorVolumeEventHandler is to handle cstor volume related events.
func (c *CStorVolumeController) cStorVolumeEventHandler(operation common.QueueOperation, cStorVolumeGot *apis.CStorVolume) (common.CStorVolumeStatus, error) {
	glog.Infof("%v event received for volume : %v ", operation, cStorVolumeGot.Name)
	switch operation {
	case common.QOpAdd:
		// CheckValidVolume is to check if volume attributes are correct.
		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			return common.CVStatusOffline, err
		}

		err = volume.CreateVolume(cStorVolumeGot)
		if err != nil {
			return common.CVStatusFailed, err
		}
		break

	case common.QOpModify:
		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			return common.CVStatusInvalid, err
		}

		err = volume.CreateVolume(cStorVolumeGot)
		if err != nil {
			return common.CVStatusFailed, err
		}
		break

	case common.QOpDestroy:
		return common.CVStatusIgnore, nil
	}

	return common.CVStatusIgnore, nil
}

// enqueueCstorVolume takes a CStorVolume resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorVolumes.
func (c *CStorVolumeController) enqueueCStorVolume(obj *apis.CStorVolume, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		glog.Errorf("Failed to enqueue %v operation for CStorVolume resource %s", q.Operation, q.Key)
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

// getVolumeResource returns object corresponding to the resource key
func (c *CStorVolumeController) getVolumeResource(key string) (*apis.CStorVolume, error) {
	// Convert the key(namespace/name) string into a distinct name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s, err: %v", key, err))
		return nil, err
	}

	if len(namespace) == 0 {
		namespace = string(common.DefaultNameSpace)
	}

	cStorVolumeGot, err := c.clientset.OpenebsV1alpha1().CStorVolumes(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorVolume resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorVolumeGot '%s' in work queue no longer exists", key))
			return nil, err
		}

		return nil, err
	}
	return cStorVolumeGot, nil
}

// IsValidCStorVolumeMgmt is to check if the volume request is for particular pod/application.
func IsValidCStorVolumeMgmt(cStorVolume *apis.CStorVolume) bool {
	if os.Getenv(string(common.OpenEBSIOCStorVolumeID)) == string(cStorVolume.UID) {
		glog.V(2).Infof("Right watcher for the cstor volume resource with id : %s", cStorVolume.UID)
		return true
	}
	glog.V(2).Infof("Wrong watcher for the cstor volume resource with id : %s", cStorVolume.UID)
	return false
}

// IsDestroyEvent is to check if the call is for cStorVolume destroy.
func IsDestroyEvent(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.ObjectMeta.DeletionTimestamp != nil {
		glog.Infof("CStor volume destroy event for volume : %s", cStorVolume.Name)
		return true
	}
	glog.Infof("CStor volume modify event for volume : %s", cStorVolume.Name)
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolume object.
func IsOnlyStatusChange(oldCStorVolume, newCStorVolume *apis.CStorVolume) bool {
	if reflect.DeepEqual(oldCStorVolume.Spec, newCStorVolume.Spec) &&
		!reflect.DeepEqual(oldCStorVolume.Status, newCStorVolume.Status) {
		glog.Infof("Only status changed for cstor volume : %s", newCStorVolume.Name)
		return true
	}
	glog.Infof("No status changed for cstor volume : %s", newCStorVolume.Name)
	return false
}

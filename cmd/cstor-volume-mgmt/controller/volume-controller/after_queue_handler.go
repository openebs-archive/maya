package volumecontroller

import (
	"fmt"
	"os"
	"reflect"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorVolumeUpdated resource
// with the current status of the resource.
func (c *CStorVolumeController) syncHandler(key string, operation common.QueueOperation) error {
	glog.Infof("at sync handler")
	cStorVolumeGot, err := c.getVolumeResource(key)
	if err != nil {
		return err
	}
	status, err := c.cStorVolumeEventHandler(operation, cStorVolumeGot)
	if status == common.CVStatusIgnore {
		return nil
	}
	cStorVolumeGot.Status.Phase = string(status)
	if err != nil {
		_, err := c.clientset.OpenebsV1alpha1().CStorVolumes().Update(cStorVolumeGot)
		if err != nil {
			return err
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorVolumes().Update(cStorVolumeGot)
	if err != nil {
		return err
	}
	return nil
}

// cStorVolumeEventHandler is to handle cstor volume related events.
func (c *CStorVolumeController) cStorVolumeEventHandler(operation common.QueueOperation, cStorVolumeGot *apis.CStorVolume) (common.CStorVolumeStatus, error) {
	volume.FileOperatorVar = util.RealFileOperator{}
	volume.UnixSockVar = util.RealUnixSock{}
	glog.Infof("%v event received for volume : %v ", operation, cStorVolumeGot.Spec.VolumeName)
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
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

// getVolumeResource returns object corresponding to the resource key
func (c *CStorVolumeController) getVolumeResource(key string) (*apis.CStorVolume, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s, err: %v", key, err))
		return nil, nil
	}

	cStorVolumeGot, err := c.clientset.OpenebsV1alpha1().CStorVolumes().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorVolume resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorVolumeGot '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return cStorVolumeGot, nil
}

// IsValidCStorVolumeMgmt is to check if the volume request is for particular pod/application.
func IsValidCStorVolumeMgmt(cStorVolume *apis.CStorVolume) bool {
	if os.Getenv("OPENEBS_IO_CSTOR_VOLUME_ID") == string(cStorVolume.ObjectMeta.UID) {
		glog.V(2).Infof("right sidecar for the cstor volume with id : %s", cStorVolume.ObjectMeta.UID)
		return true
	}
	glog.V(2).Infof("wrong sidecar for the cstor volume with id : %s", cStorVolume.ObjectMeta.UID)
	return false
}

// IsDestroyEvent is to check if the call is for cStorVolume destroy.
func IsDestroyEvent(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.ObjectMeta.DeletionTimestamp != nil {
		glog.V(2).Infof("cstor volume destroy event for volume : %s", cStorVolume.Spec.VolumeName)
		return true
	}
	glog.V(2).Infof("cstor volume modify event for volume : %s", cStorVolume.Spec.VolumeName)
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolume object.
func IsOnlyStatusChange(oldCStorVolume, newCStorVolume *apis.CStorVolume) bool {
	if reflect.DeepEqual(oldCStorVolume.Spec, newCStorVolume.Spec) &&
		!reflect.DeepEqual(oldCStorVolume.Status, newCStorVolume.Status) {
		glog.V(2).Infof("only status changed for cstor volume : %s", newCStorVolume.Spec.VolumeName)
		return true
	}
	glog.V(2).Infof("no status changed for cstor volume : %s", newCStorVolume.Spec.VolumeName)
	return false
}

// IsInitStatus is to check if the status of cStorVolume object is `init`.
func IsInitStatus(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.Status.Phase == string(common.CVStatusInit) {
		return true
	}
	return false
}

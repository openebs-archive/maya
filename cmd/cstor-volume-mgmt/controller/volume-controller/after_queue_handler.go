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
func (c *CStorVolumeController) syncHandler(key, operation string) error {
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
func (c *CStorVolumeController) cStorVolumeEventHandler(operation string, cStorVolumeGot *apis.CStorVolume) (common.CStorVolumeStatus, error) {
	volume.RunnerVar = util.RealRunner{}
	volume.FileOperatorVar = util.RealFileOperator{}
	volume.UnixSockVar = util.RealUnixSock{}
	switch operation {
	case "add":
		glog.Info("added event")
		// CheckValidVolume is to check if volume attributes are correct.
		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			return common.CVStatusOffline, err
		}

		err = volume.CreateVolume(cStorVolumeGot)
		if err != nil {
			return "", err
		}
		break

	case "modify":
		glog.Info("modify event") //ignoring as of now

		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			return "", err
		}

		err = volume.CreateVolume(cStorVolumeGot)
		if err != nil {
			return "", err
		}
		break

	case "destroy":
		glog.Info("destroy event")
		return common.CVStatusIgnore, nil
	}

	return common.CVStatusIgnore, nil
}

// enqueueCstorVolume takes a CStorVolume resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorVolumes.
func (c *CStorVolumeController) enqueueCStorVolume(obj interface{}, q common.QueueLoad) {
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
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
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

// removeFinalizer is to remove finalizer of cstorvolume resource.
func (c *CStorVolumeController) removeFinalizer(cStorVolumeGot *apis.CStorVolume) error {
	if len(cStorVolumeGot.Finalizers) > 0 {
		cStorVolumeGot.Finalizers = []string{}
	}
	_, err := c.clientset.OpenebsV1alpha1().CStorVolumes().Update(cStorVolumeGot)
	if err != nil {
		return err
	}
	return nil
}

// IsRightCStorVolumeMgmt is to check if the volume request is for particular pod/application.
func IsRightCStorVolumeMgmt(cStorVolume *apis.CStorVolume) bool {
	if os.Getenv("cstorid") == string(cStorVolume.ObjectMeta.UID) {
		glog.Infof("right sidecar")
		return true
	}
	glog.Infof("wrong sidecar")
	return false
}

// IsDestroyEvent is to check if the call is for cStorVolume destroy.
func IsDestroyEvent(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.ObjectMeta.DeletionTimestamp != nil {
		glog.Infof("cstor destroy event")
		return true
	}
	glog.Infof("cstor modify event")
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolume object.
func IsOnlyStatusChange(oldCStorVolume, newCStorVolume *apis.CStorVolume) bool {
	if reflect.DeepEqual(oldCStorVolume.Spec, newCStorVolume.Spec) &&
		!reflect.DeepEqual(oldCStorVolume.Status, newCStorVolume.Status) {
		glog.Infof("only status change")
		return true
	}
	glog.Infof("not status change")
	return false
}

// IsInitStatus is to check if the status of cStorVolume object is `init`.
func IsInitStatus(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.Status.Phase == string(common.CVStatusInit) {
		return true
	}
	return false
}

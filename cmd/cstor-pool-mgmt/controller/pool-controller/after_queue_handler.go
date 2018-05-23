package poolcontroller

import (
	"fmt"
	"os"
	"reflect"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorPoolUpdated resource
// with the current status of the resource.
func (c *CStorPoolController) syncHandler(key, operation string) error {
	glog.Infof("at sync handler")
	cStorPoolGot, err := c.getPoolResource(key)
	if err != nil {
		return err
	}
	status, err := c.cStorPoolEventHandler(operation, cStorPoolGot)
	if status == common.StatusIgnore {
		return nil
	}
	cStorPoolGot.Status.Phase = status
	if err != nil {
		_, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(cStorPoolGot)
		if err != nil {
			return err
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorPools().Update(cStorPoolGot)
	if err != nil {
		return err
	}
	return nil
}

// cStorPoolEventHandler is to handle cstor pool related events.
func (c *CStorPoolController) cStorPoolEventHandler(operation string, cStorPoolGot *apis.CStorPool) (string, error) {
	pool.RunnerVar = util.RealRunner{}
	switch operation {
	case "add":
		glog.Info("added event")
		// CheckValidPool is to check if pool attributes are correct.
		err := pool.CheckValidPool(cStorPoolGot)
		if err != nil {
			return common.StatusOffline, err
		}

		// ImportPool is to try importing pool.
		err = pool.ImportPool(cStorPoolGot)
		if err == nil {
			glog.Infof("Import Pool successful")
			return common.StatusOnline, nil
		}

		// IsInitStatus is to check if initial status of cstorpool object is `init`.
		if IsInitStatus(cStorPoolGot) {
			// LabelClear is to clear pool label
			err = pool.LabelClear(cStorPoolGot.Spec.Disks.DiskList)
			if err != nil {
				glog.Infof("Unable to clear pool labels : %v", err.Error())
			}

			// CreatePool is to create cstor pool.
			err = pool.CreatePool(cStorPoolGot)
			if err != nil {
				return common.StatusOffline, err
			}
			glog.Infof("Pool creation successful")
			return common.StatusOnline, nil
		}
		break

	case "modify":
		glog.Info("modify event") //ignoring as of now
		break

	case "destroy":
		glog.Info("destroy event")
		// DeletePool is to delete cstor pool.
		err := pool.DeletePool("cstor-" + string(cStorPoolGot.ObjectMeta.UID))
		if err != nil {
			return common.StatusDeletionFailed, err
		}
		// removeFinalizer is to remove finalizer of cStorPool resource.
		err = c.removeFinalizer(cStorPoolGot)
		if err != nil {
			return common.StatusOffline, err
		}
		return common.StatusIgnore, nil
	}

	return common.StatusIgnore, nil
}

// enqueueCstorPool takes a CStorPool resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorPools.
func (c *CStorPoolController) enqueueCStorPool(obj interface{}, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

// getPoolResource returns object corresponding to the resource key
func (c *CStorPoolController) getPoolResource(key string) (*apis.CStorPool, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	cStorPoolGot, err := c.clientset.OpenebsV1alpha1().CStorPools().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorPoolGot '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return cStorPoolGot, nil
}

// removeFinalizer is to remove finalizer of cstorpool resource.
func (c *CStorPoolController) removeFinalizer(cStorPoolGot *apis.CStorPool) error {
	if len(cStorPoolGot.Finalizers) > 0 {
		cStorPoolGot.Finalizers = []string{}
	}
	_, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(cStorPoolGot)
	if err != nil {
		return err
	}
	return nil
}

// IsRightCStorPoolMgmt is to check if the pool request is for particular pod/application.
func IsRightCStorPoolMgmt(cStorPool *apis.CStorPool) bool {
	if os.Getenv("cstorid") == string(cStorPool.ObjectMeta.UID) {
		glog.Infof("right sidecar")
		return true
	}
	glog.Infof("wrong sidecar")
	return false
}

// IsDestroyEvent is to check if the call is for cStorPool destroy.
func IsDestroyEvent(cStorPool *apis.CStorPool) bool {
	if cStorPool.ObjectMeta.DeletionTimestamp != nil {
		glog.Infof("cstor destroy event")
		return true
	}
	glog.Infof("cstor modify event")
	return false
}

// IsOnlyStatusChange is to check only status change of cStorPool object.
func IsOnlyStatusChange(oldCStorPool, newCStorPool *apis.CStorPool) bool {
	if reflect.DeepEqual(oldCStorPool.Spec, newCStorPool.Spec) &&
		!reflect.DeepEqual(oldCStorPool.Status, newCStorPool.Status) {
		glog.Infof("only status change")
		return true
	}
	glog.Infof("not status change")
	return false
}

// IsInitStatus is to check if the status of cStorPool object is `init`.
func IsInitStatus(cStorPool *apis.CStorPool) bool {
	if cStorPool.Status.Phase == common.StatusInit {
		return true
	}
	return false
}

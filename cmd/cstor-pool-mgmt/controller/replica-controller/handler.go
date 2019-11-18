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
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	"github.com/openebs/maya/pkg/debug"
	merrors "github.com/openebs/maya/pkg/errors/v1alpha1"
	pkg_errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	v130 = "1.3.0"
)

type upgradeParams struct {
	cvr    *apis.CStorVolumeReplica
	client clientset.Interface
}

type upgradeFunc func(u *upgradeParams) (*apis.CStorVolumeReplica, error)

var (
	upgradeMap = map[string]upgradeFunc{
		"1.0.0": setReplicaID,
		"1.1.0": setReplicaID,
		"1.2.0": setReplicaID,
	}
)

// CVRPatch struct represent the struct used to patch
// the cvr object
type CVRPatch struct {
	// Op defines the operation
	Op string `json:"op"`
	// Path defines the key path
	// eg. for
	// {
	//  	"Name": "openebs"
	//	    Category: {
	//		  "Inclusive": "v1",
	//		  "Rank": "A"
	//	     }
	// }
	// The path of 'Inclusive' would be
	// "/Name/Category/Inclusive"
	Path  string `json:"path"`
	Value string `json:"value"`
}

// syncHandler handles CVR changes based on the provided
// operation. It reconciles desired state of CVR with the
// actual state.
//
// Finally, it updates CVR Status
func (c *CStorVolumeReplicaController) syncHandler(
	key string,
	operation common.QueueOperation,
) error {
	cvrGot, err := c.getVolumeReplicaResource(key)
	if err != nil {
		return err
	}
	if cvrGot == nil {
		return merrors.Errorf(
			"failed to reconcile cvr {%s}: object not found",
			key,
		)
	}
	cvrGot, err = c.populateVersion(cvrGot)
	if err != nil {
		klog.Errorf("failed to add versionDetails to cvr %s:%s", cvrGot.Name, err.Error())
		c.recorder.Event(
			cvrGot,
			corev1.EventTypeWarning,
			"FailedPopulate",
			fmt.Sprintf("Failed to add current version: %s", err.Error()),
		)
		return nil
	}
	cvrGot, err = c.reconcileVersion(cvrGot)
	if err != nil {
		klog.Errorf("failed to upgrade cvr %s:%s", cvrGot.Name, err.Error())
		c.recorder.Event(
			cvrGot,
			corev1.EventTypeWarning,
			"FailedUpgrade",
			fmt.Sprintf("Failed to upgrade cvr to %s version: %s",
				cvrGot.VersionDetails.Desired,
				err.Error(),
			),
		)
		cvrGot.VersionDetails.Status.SetErrorStatus(
			"Failed to reconcile cvr version",
			err,
		)
		_, err = c.clientset.OpenebsV1alpha1().
			CStorVolumeReplicas(cvrGot.Namespace).Update(cvrGot)
		if err != nil {
			klog.Errorf("failed to update versionDetails status for cvr %s:%s", cvrGot.Name, err.Error())
		}
		return nil
	}
	status, err := c.cVREventHandler(operation, cvrGot)
	if status == "" {
		// TODO
		// need to rethink on this logic !!
		// status holds more importance than error
		return nil
	}
	cvrGot.Status.LastUpdateTime = metav1.Now()
	if cvrGot.Status.Phase != apis.CStorVolumeReplicaPhase(status) {
		cvrGot.Status.LastTransitionTime = cvrGot.Status.LastUpdateTime
		// set phase based on received status
		cvrGot.Status.Phase = apis.CStorVolumeReplicaPhase(status)
	}

	// need to update cvr before returning this error
	if err != nil {
		if debug.EI.IsCVRUpdateErrorInjected() {
			return merrors.Errorf("CVR update error via injection")
		}
		_, err1 := c.clientset.
			OpenebsV1alpha1().
			CStorVolumeReplicas(cvrGot.Namespace).
			Update(cvrGot)
		if err1 != nil {
			return merrors.Wrapf(
				err,
				"failed to reconcile cvr {%s}: failed to update cvr with phase {%s}: {%s}",
				key,
				cvrGot.Status.Phase,
				err1.Error(),
			)
		}
		return merrors.Wrapf(err, "failed to reconcile cvr {%s}", key)
	}

	// Synchronize cstor volume total allocated and
	// used capacity fields on CVR object.
	// Any kind of sync activity should be done from here.
	c.syncCvr(cvrGot)

	_, err = c.clientset.
		OpenebsV1alpha1().
		CStorVolumeReplicas(cvrGot.Namespace).
		Update(cvrGot)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to reconcile cvr {%s}: failed to update cvr with phase {%s}",
			key,
			cvrGot.Status.Phase,
		)
	}

	klog.V(4).Infof(
		"cvr {%s} reconciled successfully with current phase being {%s}",
		key,
		cvrGot.Status.Phase,
	)
	return nil
}

func (c *CStorVolumeReplicaController) cVREventHandler(
	operation common.QueueOperation,
	cvrObj *apis.CStorVolumeReplica,
) (string, error) {

	err := volumereplica.CheckValidVolumeReplica(cvrObj)
	if err != nil {
		c.recorder.Event(
			cvrObj,
			corev1.EventTypeWarning,
			string(common.FailureValidate),
			string(common.MessageResourceFailValidate),
		)
		return string(apis.CVRStatusOffline), err
	}

	// PoolNameHandler tries to get pool name and blocks for
	// particular number of attempts.
	var noOfAttempts = 2
	if !common.PoolNameHandler(cvrObj, noOfAttempts) {
		return string(apis.CVRStatusOffline), merrors.New("pool not found")
	}

	// cvr is created at zfs in the form poolname/volname
	fullVolName :=
		volumereplica.PoolNameFromCVR(cvrObj) + "/" +
			cvrObj.Labels["cstorvolume.openebs.io/name"]

	switch operation {
	case common.QOpAdd:
		klog.Infof(
			"will process add event for cvr {%s} as volume {%s}",
			cvrObj.Name,
			fullVolName,
		)

		status, err := c.cVRAddEventHandler(cvrObj, fullVolName)
		return status, err

	case common.QOpDestroy:
		klog.Infof(
			"will process delete event for cvr {%s} as volume {%s}",
			cvrObj.Name,
			fullVolName,
		)

		err := volumereplica.DeleteVolume(fullVolName)
		if err != nil {
			c.recorder.Event(
				cvrObj,
				corev1.EventTypeWarning,
				string(common.FailureDestroy),
				string(common.MessageResourceFailDestroy),
			)
			return string(apis.CVRStatusDeletionFailed), err
		}

		err = c.removeFinalizer(cvrObj)
		if err != nil {
			c.recorder.Event(
				cvrObj,
				corev1.EventTypeWarning,
				string(common.FailureRemoveFinalizer),
				string(common.MessageResourceFailDestroy),
			)
			return string(apis.CVRStatusDeletionFailed), err
		}

		return "", nil

	case common.QOpModify:
		fallthrough

	case common.QOpSync:
		klog.V(4).Infof(
			"will process sync event for cvr {%s} as volume {%s}",
			cvrObj.Name,
			operation,
		)
		return c.getCVRStatus(cvrObj)
	}

	klog.Errorf(
		"failed to handle event for cvr {%s}: operation {%s} not supported",
		cvrObj.Name,
		string(operation),
	)
	return string(apis.CVRStatusInvalid), nil
}

// removeFinalizer removes finalizers present in
// CVR resource
func (c *CStorVolumeReplicaController) removeFinalizer(
	cvrObj *apis.CStorVolumeReplica,
) error {
	cvrPatch := []CVRPatch{
		CVRPatch{
			Op:   "remove",
			Path: "/metadata/finalizers",
		},
	}

	cvrPatchBytes, err := json.Marshal(cvrPatch)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to remove finalizers from cvr {%s}",
			cvrObj.Name,
		)
	}

	_, err = c.clientset.
		OpenebsV1alpha1().
		CStorVolumeReplicas(cvrObj.Namespace).
		Patch(cvrObj.Name, types.JSONPatchType, cvrPatchBytes)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to remove finalizers from cvr {%s}",
			cvrObj.Name,
		)
	}

	klog.Infof("finalizers removed successfully from cvr {%s}", cvrObj.Name)
	return nil
}

func (c *CStorVolumeReplicaController) cVRAddEventHandler(
	cVR *apis.CStorVolumeReplica,
	fullVolName string,
) (string, error) {
	// lock is to synchronize pool and volumereplica. Until certain pool related
	// operations are over, the volumereplica threads will be held.
	common.SyncResources.Mux.Lock()
	if common.SyncResources.IsImported {
		common.SyncResources.Mux.Unlock()
		// To check if volume is already imported with pool.
		importedFlag := common.CheckForInitialImportedPoolVol(
			common.InitialImportedPoolVol,
			fullVolName,
		)
		if importedFlag && !IsEmptyStatus(cVR) {
			klog.Infof(
				"CStorVolumeReplica %v is already imported",
				string(cVR.ObjectMeta.UID),
			)

			c.recorder.Event(
				cVR,
				corev1.EventTypeNormal,
				string(common.SuccessImported),
				string(common.MessageResourceImported),
			)
			return string(apis.CVRStatusOnline), nil
		}
	} else {
		common.SyncResources.Mux.Unlock()
	}

	// Below block will be useful when the only cstor-pool-mgmt gets restarted
	// then it is required to cross-check whether the volume exists or not.
	existingvol, _ := volumereplica.GetVolumes()
	if common.CheckIfPresent(existingvol, fullVolName) {
		klog.Warningf(
			"CStorVolumeReplica %v is already present",
			string(cVR.GetUID()),
		)
		c.recorder.Event(
			cVR,
			corev1.EventTypeWarning,
			string(common.AlreadyPresent),
			string(common.MessageResourceAlreadyPresent),
		)
		// If the volume already present then return the cvr status as online
		return string(apis.CVRStatusOnline), nil
	}

	// Setting quorum to true for newly creating Volumes.
	var quorum = true
	if IsRecreateStatus(cVR) {
		klog.Infof(
			"Pool is recreated hence creating the volumes by setting off the quorum property",
		)
		quorum = false
	}

	// IsEmptyStatus is to check if initial status of cVR object is empty.
	if IsEmptyStatus(cVR) || IsInitStatus(cVR) || IsRecreateStatus(cVR) {
		// We should generate replicaID for new volume only
		if IsEmptyStatus(cVR) || IsInitStatus(cVR) {
			if err := volumereplica.GenerateReplicaID(cVR); err != nil {
				klog.Errorf("cVR ReplicaID creation failure: %v", err.Error())
				return string(apis.CVRStatusOffline), err
			}
		}
		if len(cVR.Spec.ReplicaID) == 0 {
			return string(apis.CVRStatusOffline), merrors.New("ReplicaID is not set")
		}

		err := volumereplica.CreateVolumeReplica(cVR, fullVolName, quorum)
		if err != nil {
			klog.Errorf("cVR creation failure: %v", err.Error())
			return string(apis.CVRStatusOffline), err
		}
		c.recorder.Event(
			cVR,
			corev1.EventTypeNormal,
			string(common.SuccessCreated),
			string(common.MessageResourceCreated),
		)
		klog.Infof(
			"cVR creation successful: %v, %v",
			cVR.ObjectMeta.Name,
			string(cVR.GetUID()),
		)
		return string(apis.CVRStatusOnline), nil
	}
	return string(apis.CVRStatusOffline),
		fmt.Errorf(
			"VolumeReplica offline: %v, %v",
			cVR.Name,
			cVR.Labels["cstorvolume.openebs.io/name"],
		)
}

// getVolumeReplicaResource returns object corresponding to the resource key
func (c *CStorVolumeReplicaController) getVolumeReplicaResource(
	key string,
) (*apis.CStorVolumeReplica, error) {
	// Convert the key(namespace/name) string into a distinct name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	cStorVolumeReplicaUpdated, err := c.clientset.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(
				fmt.Errorf(
					"cStorVolumeReplicaUpdated '%s' in work queue no longer exists",
					key,
				),
			)
			return nil, nil
		}
		return nil, err
	}
	return cStorVolumeReplicaUpdated, nil
}

// IsRightCStorVolumeReplica is to check if the cvr
// request is for particular pod/application.
func IsRightCStorVolumeReplica(cVR *apis.CStorVolumeReplica) bool {
	if strings.TrimSpace(string(cVR.ObjectMeta.Labels["cstorpool.openebs.io/uid"])) != "" {
		return os.Getenv(string(common.OpenEBSIOCStorID)) ==
			string(cVR.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
	}
	if strings.TrimSpace(string(cVR.ObjectMeta.Labels["cstorpoolinstance.openebs.io/uid"])) != "" {
		return os.Getenv(string(common.OpenEBSIOCSPIID)) ==
			string(cVR.ObjectMeta.Labels["cstorpoolinstance.openebs.io/uid"])
	}
	return false
}

// IsDestroyEvent is to check if the call is for CStorVolumeReplica destroy.
func IsDestroyEvent(cVR *apis.CStorVolumeReplica) bool {
	if cVR.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolumeReplica object.
func IsOnlyStatusChange(oldCVR, newCVR *apis.CStorVolumeReplica) bool {
	if reflect.DeepEqual(oldCVR.Spec, newCVR.Spec) &&
		!reflect.DeepEqual(oldCVR.Status, newCVR.Status) {
		return true
	}
	return false
}

// IsDeletionFailedBefore flags if status of
// cvr is CVRStatusDeletionFailed
func IsDeletionFailedBefore(cvrObj *apis.CStorVolumeReplica) bool {
	return cvrObj.Status.Phase == apis.CVRStatusDeletionFailed
}

// IsOnlineStatus is to check if the status of cStorVolumeReplica object is
// Healthy.
func IsOnlineStatus(cVR *apis.CStorVolumeReplica) bool {
	if string(cVR.Status.Phase) == string(apis.CVRStatusOnline) {
		klog.Infof("cVR Healthy status: %v", string(cVR.ObjectMeta.UID))
		return true
	}
	klog.Infof(
		"cVR '%s': uid '%s': phase '%s': is_healthy_status: false",
		string(cVR.ObjectMeta.Name),
		string(cVR.ObjectMeta.UID),
		cVR.Status.Phase,
	)
	return false
}

// IsEmptyStatus is to check if the status of cStorVolumeReplica object is empty.
func IsEmptyStatus(cVR *apis.CStorVolumeReplica) bool {
	if string(cVR.Status.Phase) == string(apis.CVRStatusEmpty) {
		klog.Infof("cVR empty status: %v", string(cVR.ObjectMeta.UID))
		return true
	}
	klog.Infof(
		"cVR '%s': uid '%s': phase '%s': is_empty_status: false",
		string(cVR.ObjectMeta.Name),
		string(cVR.ObjectMeta.UID),
		cVR.Status.Phase,
	)
	return false
}

// IsInitStatus is to check if the status of cStorVolumeReplica object is pending.
func IsInitStatus(cVR *apis.CStorVolumeReplica) bool {
	if string(cVR.Status.Phase) == string(apis.CVRStatusInit) {
		klog.Infof("cVR pending: %v", string(cVR.ObjectMeta.UID))
		return true
	}
	klog.V(4).Infof("Not pending status: %v", string(cVR.ObjectMeta.UID))
	return false
}

// IsRecreateStatus is to check if the status of cStorVolumeReplica object is
// in recreated state.
func IsRecreateStatus(cVR *apis.CStorVolumeReplica) bool {
	if string(cVR.Status.Phase) == string(apis.CVRStatusRecreate) {
		klog.Infof("cVR Recreate: %v", string(cVR.ObjectMeta.UID))
		return true
	}
	klog.V(4).Infof("Not Recreate status: %v", string(cVR.ObjectMeta.UID))
	return false
}

//  getCVRStatus is a wrapper that fetches the status of cstor volume.
func (c *CStorVolumeReplicaController) getCVRStatus(
	cVR *apis.CStorVolumeReplica,
) (string, error) {
	volumeName, err := volumereplica.GetVolumeName(cVR)
	if err != nil {
		return "", fmt.Errorf("unable to get volume name:%s", err.Error())
	}
	poolStatus, err := volumereplica.Status(volumeName)
	if err != nil {
		// ToDO : Put error in event recorder
		c.recorder.Event(
			cVR,
			corev1.EventTypeWarning,
			string(common.FailureStatusSync),
			string(common.MessageResourceFailStatusSync),
		)
		return "", err
	}
	return poolStatus, nil
}

// syncCvr updates field on CVR object after fetching the values from zfs utility.
func (c *CStorVolumeReplicaController) syncCvr(cvr *apis.CStorVolumeReplica) {
	// Get the zfs volume name corresponding to this cvr.
	volumeName, err := volumereplica.GetVolumeName(cvr)
	if err != nil {
		klog.Errorf("Unable to sync CVR capacity: %v", err)
		c.recorder.Event(
			cvr,
			corev1.EventTypeWarning,
			string(common.FailureCapacitySync),
			string(common.MessageResourceFailCapacitySync),
		)
	}
	// Get capacity of the volume.
	capacity, err := volumereplica.Capacity(volumeName)
	if err != nil {
		klog.Errorf("Unable to sync CVR capacity: %v", err)
		c.recorder.Event(
			cvr,
			corev1.EventTypeWarning,
			string(common.FailureCapacitySync),
			string(common.MessageResourceFailCapacitySync),
		)
	} else {
		cvr.Status.Capacity = *capacity
	}
}

func (c *CStorVolumeReplicaController) reconcileVersion(cvr *apis.CStorVolumeReplica) (
	*apis.CStorVolumeReplica, error,
) {
	var err error
	// the below code uses deep copy to have the state of object just before
	// any update call is done so that on failure the last state object can be returned
	if cvr.VersionDetails.Status.Current != cvr.VersionDetails.Desired {
		if !apis.IsCurrentVersionValid(cvr.VersionDetails.Status.Current) {
			return cvr, pkg_errors.Errorf("invalid current version %s", cvr.VersionDetails.Status.Current)
		}
		if !apis.IsDesiredVersionValid(cvr.VersionDetails.Desired) {
			return cvr, pkg_errors.Errorf("invalid desired version %s", cvr.VersionDetails.Desired)
		}
		cvrObj := cvr.DeepCopy()
		if cvrObj.VersionDetails.Status.State != apis.ReconcileInProgress {
			cvrObj.VersionDetails.Status.SetInProgressStatus()
			cvrObj, err = c.clientset.OpenebsV1alpha1().
				CStorVolumeReplicas(cvrObj.Namespace).Update(cvrObj)
			if err != nil {
				return cvr, err
			}
		}
		path := strings.Split(cvrObj.VersionDetails.Status.Current, "-")[0]
		u := &upgradeParams{
			cvr:    cvrObj,
			client: c.clientset,
		}
		// Get upgrade function for corresponding path, if path does not
		// exits then no upgrade is required and funcValue will be nil.
		funcValue := upgradeMap[path]
		if funcValue != nil {
			cvrObj, err = funcValue(u)
			if err != nil {
				return cvrObj, err
			}
		}
		cvr = cvrObj.DeepCopy()
		cvrObj.VersionDetails.SetSuccessStatus()
		cvrObj, err = c.clientset.OpenebsV1alpha1().
			CStorVolumeReplicas(cvrObj.Namespace).Update(cvrObj)
		if err != nil {
			return cvr, err
		}
		return cvrObj, nil
	}
	return cvr, nil
}

// populateVersion assigns VersionDetails for old cvr object
func (c *CStorVolumeReplicaController) populateVersion(cvr *apis.CStorVolumeReplica) (
	*apis.CStorVolumeReplica, error,
) {
	v := cvr.Labels[string(apis.OpenEBSVersionKey)]
	// 1.3.0 onwards new CVR will have the field populated during creation
	if v < v130 && cvr.VersionDetails.Status.Current == "" {
		cvrObj := cvr.DeepCopy()
		cvrObj.VersionDetails.Status.Current = v
		cvrObj.VersionDetails.Desired = v
		cvrObj, err := c.clientset.OpenebsV1alpha1().CStorVolumeReplicas(cvrObj.Namespace).
			Update(cvrObj)

		if err != nil {
			return cvr, err
		}
		klog.Infof("Version %s added on cvr %s", v, cvrObj.Name)
		return cvrObj, nil
	}
	return cvr, nil
}

// setReplicaID sets the replica_id if not present for old cvrs when
// they are upgraded to version 1.3.0 or above.
func setReplicaID(u *upgradeParams) (*apis.CStorVolumeReplica, error) {
	cvr := u.cvr
	cvrObj := cvr.DeepCopy()
	err := volumereplica.GetAndUpdateReplicaID(cvrObj)
	if err != nil {
		return cvr, err
	}
	cvrObj, err = u.client.OpenebsV1alpha1().
		CStorVolumeReplicas(cvrObj.Namespace).Update(cvrObj)
	if err != nil {
		return cvr, err
	}
	return cvrObj, nil
}

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
	"time"

	pkg_errors "github.com/pkg/errors"

	"strings"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	"github.com/openebs/maya/pkg/client/k8s"
	cstorvolume_v1alpha1 "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type upgradeParams struct {
	cv     *apis.CStorVolume
	client clientset.Interface
}

type upgradeFunc func(u *upgradeParams) (*apis.CStorVolume, error)

var (
	v130       = "1.3.0"
	upgradeMap = map[string]upgradeFunc{
		"1.0.0": setDesiredRF,
		"1.1.0": setDesiredRF,
		"1.2.0": setDesiredRF,
		"1.3.0": nothing,
	}
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorVolumeUpdated
// resource with the current status of the resource.
func (c *CStorVolumeController) syncHandler(
	key string,
	operation common.QueueOperation,
) error {
	klog.V(4).Infof("Handling %v operation for resource : %s ", operation, key)
	cStorVolumeGot, err := c.getVolumeResource(key)
	if err != nil {
		return err
	}
	cStorVolumeGot, err = c.populateVersion(cStorVolumeGot)
	if err != nil {
		klog.Errorf("failed to add versionDetails to cv %s:%s", cStorVolumeGot.Name, err.Error())
		c.recorder.Event(
			cStorVolumeGot,
			corev1.EventTypeWarning,
			"FailedPopulate",
			fmt.Sprintf("Failed to add current version: %s", err.Error()),
		)
		return nil
	}
	cStorVolumeGot, err = c.reconcileVersion(cStorVolumeGot)
	if err != nil {
		klog.Errorf("failed to upgrade cv %s:%s", cStorVolumeGot.Name, err.Error())
		c.recorder.Event(
			cStorVolumeGot,
			corev1.EventTypeWarning,
			"FailedUpgrade",
			fmt.Sprintf("Failed to upgrade cv to %s version: %s",
				cStorVolumeGot.VersionDetails.Desired,
				err.Error(),
			),
		)
		cStorVolumeGot.VersionDetails.Status.SetErrorStatus(
			"Failed to reconcile cv version",
			err,
		)
		_, err = c.clientset.OpenebsV1alpha1().
			CStorVolumes(cStorVolumeGot.Namespace).Update(cStorVolumeGot)
		if err != nil {
			klog.Errorf("failed to update versionDetails status for cv %s:%s", cStorVolumeGot.Name, err.Error())
		}
		return nil
	}
	status, err := c.cStorVolumeEventHandler(operation, cStorVolumeGot)
	if status == common.CVStatusIgnore {
		return nil
	}
	cStorVolumeGot.Status.LastUpdateTime = metav1.Now()
	if cStorVolumeGot.Status.Phase != apis.CStorVolumePhase(status) {
		cStorVolumeGot.Status.LastTransitionTime = cStorVolumeGot.Status.LastUpdateTime
		cStorVolumeGot.Status.Phase = apis.CStorVolumePhase(status)
	}
	if err != nil {
		klog.Errorf(err.Error())
		klog.Infof("cStorVolume:%v, %v; Status: %v", cStorVolumeGot.Name,
			string(cStorVolumeGot.GetUID()), cStorVolumeGot.Status.Phase)
		_, err1 := c.clientset.OpenebsV1alpha1().
			CStorVolumes(cStorVolumeGot.Namespace).
			Update(cStorVolumeGot)
		if err1 != nil {
			return pkg_errors.Wrapf(
				err1,
				"failed to update cStorVolume:%v, %v; Status: %v err: %v",
				cStorVolumeGot.Name,
				string(cStorVolumeGot.GetUID()),
				cStorVolumeGot.Status.Phase,
				err,
			)
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().
		CStorVolumes(cStorVolumeGot.Namespace).
		Update(cStorVolumeGot)
	if err != nil {
		return pkg_errors.Wrapf(
			err, "failed to update cStorVolume:%v, %v; Status: %v",
			cStorVolumeGot.Name,
			string(cStorVolumeGot.GetUID()),
			cStorVolumeGot.Status.Phase,
		)
	}
	klog.V(4).Infof("cStorVolume:%v, %v; Status: %v", cStorVolumeGot.Name,
		string(cStorVolumeGot.GetUID()), cStorVolumeGot.Status.Phase)
	return nil
}

//TODO: Return status of cstorvolume and patch it on caller of below function

// cStorVolumeEventHandler is to handle cstor volume related events.
func (c *CStorVolumeController) cStorVolumeEventHandler(
	operation common.QueueOperation,
	cStorVolumeGot *apis.CStorVolume,
) (common.CStorVolumeStatus, error) {
	var eventMessage string
	var updatedCV *apis.CStorVolume
	customCVObj := cstorvolume_v1alpha1.NewForAPIObject(cStorVolumeGot)
	klog.V(4).Infof(
		"%v event received for volume : %v ",
		operation,
		cStorVolumeGot.Name,
	)
	switch operation {
	case common.QOpAdd:
		// CheckValidVolume is to check if volume attributes are correct.
		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			c.recorder.Event(
				cStorVolumeGot, corev1.EventTypeWarning,
				string(common.FailureCreate),
				fmt.Sprintf("failed to create cstorvolume validation "+
					"failed on cstorvolum error: %v", err),
			)
			return common.CVStatusInvalid, err
		}
		// Set TargetNamespace which will be used to volume-mgmt UDS to update
		//replication information
		cstorvolume_v1alpha1.TargetNamespace = cStorVolumeGot.Namespace

		err = volume.CreateVolumeTarget(cStorVolumeGot)
		if err != nil {
			return common.CVStatusError, err
		}
		// update the status capacity of cstorvolume caller of this code
		// will update in etcd
		if !customCVObj.IsResizePending() {
			cStorVolumeGot.Status.Capacity = cStorVolumeGot.Spec.Capacity
		}
		return common.CVStatusInit, nil

	case common.QOpModify:
		// Make changes here to run zrepl command and update the data
		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			c.recorder.Event(
				cStorVolumeGot, corev1.EventTypeWarning,
				string(common.FailureCreate),
				fmt.Sprintf("failed to update cstorvolume validation "+
					"failed on cstorvolume errors: %v", err),
			)
			return common.CVStatusInvalid, err
		}
		if customCVObj.IsDRFPending() {
			c.updateCStorVolumeDRF(cStorVolumeGot)
		}
		// blocking call for doing resize operation
		if customCVObj.IsResizePending() {
			if !customCVObj.IsConditionPresent(apis.CStorVolumeResizing) {
				updatedCV, err = c.addResizeConditions(cStorVolumeGot)
				if err != nil {
					return common.CVStatusIgnore, nil
				}
				cStorVolumeGot = updatedCV
			}
			_, err = c.resizeCStorVolume(cStorVolumeGot)
			if err != nil {
				klog.Errorf(
					"failed to resize cstorvolume %s from %s to %s error %v",
					cStorVolumeGot.Name,
					cStorVolumeGot.Status.Capacity.String(),
					cStorVolumeGot.Spec.Capacity.String(),
					err,
				)
				// return ignore from here so that caller does not
				// attempt to update CV again. Since resize is handled during
				// sync time also.
			}
			return common.CVStatusIgnore, nil
		}
		eventMessage = fmt.Sprintf("Ignoring changes on volume %s", cStorVolumeGot.Name)
		c.recorder.Event(
			cStorVolumeGot, corev1.EventTypeWarning,
			string(common.FailureUpdate), eventMessage,
		)

	case common.QOpPeriodicSync:
		var err error
		lastKnownPhase := cStorVolumeGot.Status.Phase
		if customCVObj.IsDRFPending() {
			c.updateCStorVolumeDRF(cStorVolumeGot)
		}
		// blocking call for doing resize operation
		if customCVObj.IsResizePending() {
			if !customCVObj.IsConditionPresent(apis.CStorVolumeResizing) {
				updatedCV, err = c.addResizeConditions(cStorVolumeGot)
				if err != nil {
					//NOTE: Only after updating CV with resize conditions
					//process resize porcess
					goto volumeStatus
				}
				cStorVolumeGot = updatedCV
			}
			// return same as previous state
			updatedCV, err = c.resizeCStorVolume(cStorVolumeGot)
			if err != nil {
				klog.Errorf(
					"failed to resize cstorvolume %s from %s to %s error %v",
					cStorVolumeGot.Name,
					cStorVolumeGot.Status.Capacity.String(),
					cStorVolumeGot.Spec.Capacity.String(),
					err,
				)
			} else {
				cStorVolumeGot = updatedCV
			}
		}
	volumeStatus:
		volStatus, err := volume.GetVolumeStatus(cStorVolumeGot)
		if err != nil {
			klog.Errorf("Error in getting volume status: %s", err.Error())
			return common.CVStatusError, nil
		} else {
			cStorVolumeGot.Status.Phase = apis.CStorVolumePhase(volStatus.Status)
			// if replicas are zero set the status as init
			if len(volStatus.ReplicaStatuses) == 0 {
				cStorVolumeGot.Status.Phase = apis.CStorVolumePhase(
					common.CVStatusInit,
				)
			}
		}
		cStorVolumeGot.Status.LastUpdateTime = metav1.Now()
		if cStorVolumeGot.Status.Phase != lastKnownPhase {
			cStorVolumeGot.Status.LastTransitionTime = cStorVolumeGot.Status.LastUpdateTime
		}

		cStorVolumeGot.Status.ReplicaStatuses = volStatus.ReplicaStatuses
		updatedCstorVolume, err := c.clientset.OpenebsV1alpha1().
			CStorVolumes(cStorVolumeGot.Namespace).
			Update(cStorVolumeGot)
		if err != nil {
			klog.Errorf("Error updating cStorVolume object: %s", err)
			return common.CVStatusIgnore, nil
		}
		// if there is no change in the phase of the cv only then create event
		if lastKnownPhase != updatedCstorVolume.Status.Phase {
			err = c.createSyncUpdateEvent(c.createEventObj(updatedCstorVolume))
			if err != nil {
				klog.Errorf("Error creating event : %s", err.Error())
			}
		}
		// Update already made above with latest status.
		// We return ignore from here so that caller does not
		// re-attempt to update status with older resource version
		return common.CVStatusIgnore, nil
	case common.QOpDestroy:
		return common.CVStatusIgnore, nil
	}

	klog.Infof(
		"Ignoring changes for volume %s for operation %v",
		cStorVolumeGot.Name,
		operation,
	)
	return common.CVStatusIgnore, nil
}

// getEventType returns the event type based on the passed CStorVolumeStatus
func getEventType(phase common.CStorVolumeStatus) string {
	// It is normal event only when phase is Running or Degraded
	if phase == common.CVStatusInit ||
		phase == common.CVStatusHealthy ||
		phase == common.CVStatusDegraded {
		return corev1.EventTypeNormal
	}
	return corev1.EventTypeWarning
}

// createEventObj creates an object of corev1.Event based on the CstorVolume
func (c *CStorVolumeController) createEventObj(
	cstorVolume *apis.CStorVolume,
) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cstorVolume.Name + "." + string(cstorVolume.Status.Phase),
			Namespace: cstorVolume.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:            string(k8s.CStorVolumeCRKK),
			APIVersion:      string(k8s.OEV1alpha1KA),
			Name:            cstorVolume.Name,
			Namespace:       cstorVolume.Namespace,
			UID:             cstorVolume.UID,
			ResourceVersion: cstorVolume.ResourceVersion,
		},
		FirstTimestamp: metav1.Time{Time: time.Now()},
		LastTimestamp:  metav1.Time{Time: time.Now()},
		Count:          1,
		Message:        fmt.Sprintf(common.EventMsgFormatter, cstorVolume.Status.Phase),
		Reason:         string(cstorVolume.Status.Phase),
		Type:           getEventType(common.CStorVolumeStatus(cstorVolume.Status.Phase)),
		Source: corev1.EventSource{
			Component: os.Getenv("POD_NAME"),
			Host:      os.Getenv("NODE_NAME"),
		},
	}
}

// createSyncUpdateEvent tries to get the eventGot if present it updates
// the lastTimestamp to current time and increases count by one,
// if absent, creates the given eventGot
func (c *CStorVolumeController) createSyncUpdateEvent(
	eventGot *corev1.Event,
) (err error) {
	client := c.kubeclientset
	event, err := client.CoreV1().
		Events(eventGot.Namespace).
		Get(eventGot.Name, metav1.GetOptions{})
	// error could be due to missing object or some other reason
	// we ignore error if it is due to missing object
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}
	// checking name as sometimes we receive an empty event object instead of nil
	if event == nil || len(event.Name) == 0 {
		// create event
		event, err = client.CoreV1().Events(eventGot.Namespace).Create(eventGot)
	} else {
		event.Count = event.Count + 1
		event.LastTimestamp = metav1.Time{Time: time.Now()}
		// update the event with increased count and new timestamp
		_, err = client.CoreV1().Events(eventGot.Namespace).Update(event)
	}
	return
}

// enqueueCstorVolume takes a CStorVolume resource and converts it into a
// namespace/name string which is then put onto the work queue. This method
// should *not* be passed resources of any type other than CStorVolumes.
func (c *CStorVolumeController) enqueueCStorVolume(
	obj *apis.CStorVolume,
	q common.QueueLoad,
) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		klog.Errorf(
			"Failed to enqueue %v operation for CStorVolume resource %s",
			q.Operation,
			q.Key,
		)
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

// getVolumeResource returns object corresponding to the resource key
func (c *CStorVolumeController) getVolumeResource(
	key string,
) (*apis.CStorVolume, error) {
	// Convert the key(namespace/name) string into a distinct name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(
			fmt.Errorf("Invalid resource key: %s, err: %v", key, err),
		)
		return nil, err
	}

	if len(namespace) == 0 {
		namespace = string(common.DefaultNameSpace)
	}

	cStorVolumeGot, err := c.clientset.OpenebsV1alpha1().
		CStorVolumes(namespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorVolume resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(
				fmt.Errorf(
					"cStorVolumeGot '%s' in work queue no longer exists",
					key,
				),
			)
			return nil, err
		}

		return nil, err
	}
	return cStorVolumeGot, nil
}

// addResizeConditions will add resize condition to cstorvolume
func (c *CStorVolumeController) addResizeConditions(
	cvObj *apis.CStorVolume) (*apis.CStorVolume, error) {
	resizeConditions := cstorvolume_v1alpha1.GetResizeCondition()
	var eventMessage string
	cvObj.Status.Conditions = cstorvolume_v1alpha1.
		Conditions(cvObj.Status.Conditions).
		AddCondition(resizeConditions)
	updatedCVObj, err := c.clientset.
		OpenebsV1alpha1().
		CStorVolumes(cvObj.Namespace).
		Update(cvObj)
	if err != nil {
		// Generate event and return
		eventMessage = fmt.Sprintf(
			"failed to update resize conditions error %v",
			err,
		)
		c.recorder.Event(
			cvObj,
			corev1.EventTypeWarning,
			string(common.FailureUpdate),
			eventMessage,
		)
		return nil, pkg_errors.New(eventMessage)
	}
	c.recorder.Event(
		cvObj,
		corev1.EventTypeNormal,
		string(common.SuccessUpdated),
		"Updated resize conditions",
	)
	return updatedCVObj, nil
}

// updateCStorVolumeDRF updates volume configurations to changes made to
// cStorVolume CR.
// If no.of entries in cStorvolume spec.ReplicaDetails.KnownReplicas is less
// than entries in status.ReplicaDetails.knownReplicas then the changes are
// identified for scaledown
// process.
// If no.of entries in cStorVolume spec.ReplicaDetails.KnownReplicas is equal to
// no.of entries in stauts.ReplicaDetails.KnownReplicas then the changes are
// identified for scaleup process(Only one process should be triggered either
// scaleup or scaledown).
// NOTE: No.of entries in spec.KnownReplicas and status.knownReplicas will vary
// only in the scaledown process else always it will be same
func (c *CStorVolumeController) updateCStorVolumeDRF(
	cStorVolume *apis.CStorVolume) {
	// make changes to copyCV instead of cStorVolume
	copyCV := cStorVolume.DeepCopy()
	var err error
	specKnownReplicaCount := len(copyCV.Spec.ReplicaDetails.KnownReplicas)
	statusKnownReplicaCount := len(copyCV.Status.ReplicaDetails.KnownReplicas)
	if specKnownReplicaCount < statusKnownReplicaCount {
		copyCV, err = c.triggerScaleDownProcess(copyCV)
	} else if specKnownReplicaCount == statusKnownReplicaCount {
		copyCV, err = c.triggerScaleUpProcess(copyCV)
	} else {
		// entries in spec.ReplicaDetails.KnownReplicas can't be greater than
		// status.ReplicaDetails.KnownReplicas(unkown changes)
		err = pkg_errors.Errorf("unkown changes are made to cStorVolume no.of "+
			"entries in spec.ReplicaDetails.knownReplicas are %d and no.of entries "+
			"in status.ReplicaDetails.KnownReplicas are %d",
			specKnownReplicaCount,
			statusKnownReplicaCount)
	}
	if err != nil {
		c.recorder.Event(cStorVolume,
			corev1.EventTypeWarning,
			string(common.FailureUpdate),
			fmt.Sprintf("failed to update desired replication factor to %d"+
				" Error: %v", cStorVolume.Spec.DesiredReplicationFactor, err,
			),
		)
		return
	}
	cStorVolume = copyCV
	c.recorder.Event(cStorVolume,
		corev1.EventTypeNormal,
		string(common.SuccessUpdated),
		fmt.Sprintf("Successfully updated the desiredReplicationFactor to %d",
			cStorVolume.Spec.DesiredReplicationFactor),
	)
}

// triggerScaleUpProcess returns error in case of any error occurred during scaleup
// process else it will return cstorvolume object
func (c *CStorVolumeController) triggerScaleUpProcess(
	cStorVolume *apis.CStorVolume) (*apis.CStorVolume, error) {
	err := volume.ExecuteDesiredReplicationFactorCommand(
		cStorVolume,
		volume.GetScaleUpCommand,
	)
	if err != nil {
		return nil, err
	}
	fileOperator := util.RealFileOperator{}
	updatedConfig := fmt.Sprintf("%s %d",
		cstorvolume_v1alpha1.DesiredReplicationFactorKey,
		cStorVolume.Spec.DesiredReplicationFactor)
	cstorvolume_v1alpha1.ConfFileMutex.Lock()
	err = fileOperator.Updatefile(cstorvolume_v1alpha1.IstgtConfPath,
		updatedConfig,
		cstorvolume_v1alpha1.DesiredReplicationFactorKey,
		0644,
	)
	cstorvolume_v1alpha1.ConfFileMutex.Unlock()
	if err != nil {
		return nil, pkg_errors.Wrapf(err, "failed to update conf file with "+
			"desiredReplicationFactor %d", cStorVolume.Spec.DesiredReplicationFactor)
	}
	return cStorVolume, nil
}

// triggerScaleDownProcess returns error in case of any error during the scaledown
// process else it will return cstorvolume object. Following steps are executed
// 1. Verify whether all the replicas are healthy other than removing replica.
// 2. If step1 is passed then update the istgt.conf file with latest
//    information.
// 3. Update cStorVolume CR(replicationFactor, ConsistencyFactor and known
//    replica list.
// 4. Trigger istgtcontrol command (istgtcontrol drf <vol_name> <value>
//    <remaining_replica_list>
// In triggerScaleDownProcess cStorVolumeAPI is used to access fields in CV
// object and cStorvolume is used to access methods of cStorVolume
func (c *CStorVolumeController) triggerScaleDownProcess(
	cStorVolumeAPI *apis.CStorVolume) (*apis.CStorVolume, error) {
	volumeStatus, err := volume.GetVolumeStatus(cStorVolumeAPI)
	if err != nil {
		return nil, err
	}
	if common.CStorVolumeStatus(volumeStatus.Status) != common.CVStatusHealthy {
		return nil, pkg_errors.Errorf("cStorvolume is not in healthy to trigger scaledown")
	}
	cStorVolume := cstorvolume_v1alpha1.NewForAPIObject(cStorVolumeAPI)
	if !cStorVolume.AreSpecReplicasHealthy(volumeStatus) {
		return nil, pkg_errors.Errorf("spec replicas are not in healthy state")
	}
	replicaID := cStorVolume.GetRemovingReplicaID()
	if replicaID == "" {
		return nil, pkg_errors.Errorf("removing replica is not present in volume status")
	}
	if cStorVolumeAPI.Spec.DesiredReplicationFactor < cStorVolumeAPI.Spec.ReplicationFactor {
		configData := cStorVolume.BuildScaleDownConfigData(replicaID)
		fileOperator := util.RealFileOperator{}
		cstorvolume_v1alpha1.ConfFileMutex.Lock()
		err = fileOperator.UpdateOrAppendMultipleLines(
			cstorvolume_v1alpha1.IstgtConfPath,
			configData,
			0644)
		cstorvolume_v1alpha1.ConfFileMutex.Unlock()
		if err != nil {
			return nil, pkg_errors.Wrapf(err, "failed to update istgt conf file %v", configData)
		}
		cStorVolumeAPI.Spec.ReplicationFactor = cStorVolumeAPI.Spec.DesiredReplicationFactor
		cStorVolumeAPI.Spec.ConsistencyFactor = (cStorVolumeAPI.Spec.ReplicationFactor)/2 + 1
		cStorVolumeAPI, err = c.clientset.
			OpenebsV1alpha1().
			CStorVolumes(cStorVolumeAPI.Namespace).
			Update(cStorVolumeAPI)
		if err != nil {
			return nil, pkg_errors.Wrapf(err, "failed to update cstorvolume")
		}
	}
	err = volume.ExecuteDesiredReplicationFactorCommand(
		cStorVolumeAPI,
		volume.GetScaleDownCommand,
	)
	if err != nil {
		return nil, err
	}
	cStorVolumeAPI.Status.ReplicaDetails.KnownReplicas =
		cStorVolumeAPI.Spec.ReplicaDetails.KnownReplicas
	cStorVolumeAPI, err = c.clientset.
		OpenebsV1alpha1().
		CStorVolumes(cStorVolumeAPI.Namespace).
		Update(cStorVolumeAPI)
	if err != nil {
		return nil, pkg_errors.Wrapf(err,
			"failed to update cstorvolume status with scaledown replica information",
		)
	}
	return cStorVolumeAPI, nil
}

// resizeCStorVolume resize the cstorvolume and if any error occurs updates the
// resize conditions of cstorvolume either with success or failure message
func (c *CStorVolumeController) resizeCStorVolume(
	cStorVolume *apis.CStorVolume) (*apis.CStorVolume, error) {
	var eventMessage string
	var err error
	isResizeSuccess := false
	copyCV := cStorVolume.DeepCopy()
	customCVObj := cstorvolume_v1alpha1.NewForAPIObject(copyCV)
	// NOTE: We are processing resize process only after updating CV resize conditions
	// so in below call there is no chance of geting new CVCondition instance
	conditionStatus := customCVObj.GetCVCondition(apis.CStorVolumeResizing)
	desiredCap := copyCV.Spec.Capacity.String()

	err = volume.ResizeTargetVolume(copyCV)
	if err != nil {
		eventMessage = fmt.Sprintf(
			"failed to resize cstorvolume from %s to %s error %v",
			cStorVolume.Status.Capacity.String(),
			desiredCap,
			err,
		)
		c.recorder.Event(copyCV, corev1.EventTypeWarning, string(common.FailureUpdate), eventMessage)
		conditionStatus.Message = eventMessage
		copyCV.Status.Conditions = cstorvolume_v1alpha1.
			Conditions(copyCV.Status.Conditions).
			UpdateCondition(conditionStatus)
	} else {
		// In success case remove resize condition
		copyCV.Status.Conditions = cstorvolume_v1alpha1.
			Conditions(copyCV.Status.Conditions).
			DeleteCondition(conditionStatus)
		copyCV.Status.Capacity = copyCV.Spec.Capacity
		isResizeSuccess = true
	}

	newCV, cvUpdateErr := c.clientset.
		OpenebsV1alpha1().
		CStorVolumes(copyCV.Namespace).
		Update(copyCV)
	if cvUpdateErr == nil && isResizeSuccess {
		eventMessage = fmt.Sprintf(
			"successfully resized volume from %s to %s",
			cStorVolume.Status.Capacity.String(),
			desiredCap,
		)
		c.recorder.Event(copyCV, corev1.EventTypeNormal, string(common.SuccessUpdated), eventMessage)
	}
	return newCV, cvUpdateErr
}

// IsValidCStorVolumeMgmt is to check if the volume request
// is for particular pod/application.
func IsValidCStorVolumeMgmt(cStorVolume *apis.CStorVolume) bool {
	if os.Getenv(string(common.OpenEBSIOCStorVolumeID)) == string(cStorVolume.UID) {
		klog.V(2).Infof(
			"Right watcher for the cstor volume resource with id : %s",
			cStorVolume.UID,
		)
		return true
	}
	klog.V(2).Infof(
		"Wrong watcher for the cstor volume resource with id : %s",
		cStorVolume.UID,
	)
	return false
}

// IsDestroyEvent is to check if the call is for cStorVolume destroy.
func IsDestroyEvent(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.ObjectMeta.DeletionTimestamp != nil {
		klog.Infof(
			"CStor volume destroy event for volume : %s",
			cStorVolume.Name,
		)
		return true
	}
	klog.Infof("CStor volume modify event for volume : %s", cStorVolume.Name)
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolume object.
func IsOnlyStatusChange(oldCStorVolume, newCStorVolume *apis.CStorVolume) bool {
	if reflect.DeepEqual(oldCStorVolume.Spec, newCStorVolume.Spec) &&
		!reflect.DeepEqual(oldCStorVolume.Status, newCStorVolume.Status) {
		klog.Infof(
			"Only status changed for cstor volume : %s",
			newCStorVolume.Name,
		)
		return true
	}
	klog.Infof("No status changed for cstor volume : %s", newCStorVolume.Name)
	return false
}

func (c *CStorVolumeController) reconcileVersion(cv *apis.CStorVolume) (*apis.CStorVolume, error) {
	var err error
	// the below code uses DeepCopy() to have the state of object just before
	// any update call is done so that on failure the last state object can be returned
	if cv.VersionDetails.Status.Current != cv.VersionDetails.Desired {
		if !apis.IsCurrentVersionValid(cv.VersionDetails.Status.Current) {
			return cv, pkg_errors.Errorf("invalid current version %s", cv.VersionDetails.Status.Current)
		}
		if !apis.IsDesiredVersionValid(cv.VersionDetails.Desired) {
			return cv, pkg_errors.Errorf("invalid desired version %s", cv.VersionDetails.Desired)
		}
		cvObject := cv.DeepCopy()
		if cv.VersionDetails.Status.State != apis.ReconcileInProgress {
			cvObject.VersionDetails.Status.SetInProgressStatus()
			cvObject, err = c.clientset.OpenebsV1alpha1().
				CStorVolumes(cvObject.Namespace).Update(cvObject)
			if err != nil {
				return cv, err
			}
		}
		path := strings.Split(cvObject.VersionDetails.Status.Current, "-")[0]
		u := &upgradeParams{
			cv:     cvObject,
			client: c.clientset,
		}
		cvObject, err = upgradeMap[path](u)
		if err != nil {
			return cvObject, err
		}
		cv = cvObject.DeepCopy()
		cvObject.VersionDetails.Status.SetSuccessStatus()
		cvObject, err = c.clientset.OpenebsV1alpha1().
			CStorVolumes(cvObject.Namespace).Update(cvObject)
		if err != nil {
			return cv, err
		}
		return cvObject, nil
	}
	return cv, nil
}

// populateVersion assigns VersionDetails for old cv object
func (c *CStorVolumeController) populateVersion(cv *apis.CStorVolume) (
	*apis.CStorVolume, error,
) {
	v := cv.Labels[string(apis.OpenEBSVersionKey)]
	// 1.3.0 onwards new CV will have the field populated during creation
	if v < v130 && cv.VersionDetails.Status.Current == "" {
		cvObj := cv.DeepCopy()
		cvObj.VersionDetails.Status.Current = v
		cvObj.VersionDetails.Desired = v
		cvObj, err := c.clientset.OpenebsV1alpha1().CStorVolumes(cvObj.Namespace).
			Update(cvObj)

		if err != nil {
			return cv, err
		}
		klog.Infof("Version %s added on cstorvolume %s", v, cvObj.Name)
		return cvObj, nil
	}
	return cv, nil
}

func setDesiredRF(u *upgradeParams) (*apis.CStorVolume, error) {
	cv := u.cv
	// Set new field DesiredReplicationFactor as ReplicationFactor
	cv.Spec.DesiredReplicationFactor = cv.Spec.ReplicationFactor
	return cv, nil
}

func nothing(u *upgradeParams) (*apis.CStorVolume, error) {
	return u.cv, nil
}

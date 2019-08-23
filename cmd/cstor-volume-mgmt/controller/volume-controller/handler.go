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

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/k8s"
	cstorvolume "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorVolumeUpdated
// resource with the current status of the resource.
func (c *CStorVolumeController) syncHandler(
	key string,
	operation common.QueueOperation,
) error {
	glog.V(4).Infof("Handling %v operation for resource : %s ", operation, key)
	cStorVolumeGot, err := c.getVolumeResource(key)
	if err != nil {
		return err
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
		glog.Errorf(err.Error())
		glog.Infof("cStorVolume:%v, %v; Status: %v", cStorVolumeGot.Name,
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
	glog.V(4).Infof("cStorVolume:%v, %v; Status: %v", cStorVolumeGot.Name,
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
	customCVObj := cstorvolume.NewForAPIObject(cStorVolumeGot)
	glog.V(4).Infof(
		"%v event received for volume : %v ",
		operation,
		cStorVolumeGot.Name,
	)
	switch operation {
	case common.QOpAdd:
		// CheckValidVolume is to check if volume attributes are correct.
		err := volume.CheckValidVolume(cStorVolumeGot)
		if err != nil {
			return common.CVStatusInvalid, err
		}

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
			return common.CVStatusInvalid, err
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
				glog.Errorf(
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
				glog.Errorf(
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
			glog.Errorf("Error in getting volume status: %s", err.Error())
			cStorVolumeGot.Status.Phase = apis.CStorVolumePhase(
				common.CVStatusError,
			)
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
			glog.Errorf("Error updating cStorVolume object: %s", err)
			return common.CVStatusIgnore, nil
		}
		// if there is no change in the phase of the cv only then create event
		if lastKnownPhase != updatedCstorVolume.Status.Phase {
			err = c.createSyncUpdateEvent(c.createEventObj(updatedCstorVolume))
			if err != nil {
				glog.Errorf("Error creating event : %s", err.Error())
			}
		}
		// Update already made above with latest status.
		// We return ignore from here so that caller does not
		// re-attempt to update status with older resource version
		return common.CVStatusIgnore, nil
	case common.QOpDestroy:
		return common.CVStatusIgnore, nil
	}

	glog.Infof(
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
		glog.Errorf(
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
	resizeConditions := cstorvolume.GetResizeCondition()
	var eventMessage string
	cvObj.Status.Conditions = cstorvolume.
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

// resizeCStorVolume resize the cstorvolume and if any error occurs updates the
// resize conditions of cstorvolume either with success or failure message
func (c *CStorVolumeController) resizeCStorVolume(
	cStorVolume *apis.CStorVolume) (*apis.CStorVolume, error) {
	var eventMessage string
	var err error
	isResizeSuccess := false
	copyCV := cStorVolume.DeepCopy()
	customCVObj := cstorvolume.NewForAPIObject(copyCV)
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
		copyCV.Status.Conditions = cstorvolume.
			Conditions(copyCV.Status.Conditions).
			UpdateCondition(conditionStatus)
	} else {
		// In success case remove resize condition
		copyCV.Status.Conditions = cstorvolume.
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
		glog.V(2).Infof(
			"Right watcher for the cstor volume resource with id : %s",
			cStorVolume.UID,
		)
		return true
	}
	glog.V(2).Infof(
		"Wrong watcher for the cstor volume resource with id : %s",
		cStorVolume.UID,
	)
	return false
}

// IsDestroyEvent is to check if the call is for cStorVolume destroy.
func IsDestroyEvent(cStorVolume *apis.CStorVolume) bool {
	if cStorVolume.ObjectMeta.DeletionTimestamp != nil {
		glog.Infof(
			"CStor volume destroy event for volume : %s",
			cStorVolume.Name,
		)
		return true
	}
	glog.Infof("CStor volume modify event for volume : %s", cStorVolume.Name)
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolume object.
func IsOnlyStatusChange(oldCStorVolume, newCStorVolume *apis.CStorVolume) bool {
	if reflect.DeepEqual(oldCStorVolume.Spec, newCStorVolume.Spec) &&
		!reflect.DeepEqual(oldCStorVolume.Status, newCStorVolume.Status) {
		glog.Infof(
			"Only status changed for cstor volume : %s",
			newCStorVolume.Name,
		)
		return true
	}
	glog.Infof("No status changed for cstor volume : %s", newCStorVolume.Name)
	return false
}

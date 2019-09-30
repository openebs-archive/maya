// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"k8s.io/klog"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// ConfFileMutex is to hold the lock while updating istgt.conf file
	ConfFileMutex = &sync.Mutex{}
	// IstgtConfPath will locate path for istgt configurations
	IstgtConfPath = "/usr/local/etc/istgt/istgt.conf"
	//DesiredReplicationFactorKey is plain text in istgt configuration file informs
	//about desired replication factor used by target
	DesiredReplicationFactorKey = "  DesiredReplicationFactor"
	//TargetNamespace holds key of env where target pod is running
	TargetNamespace = "CSTOR_TARGET_NAMESPACE"
)

const (
	//IoWaitTime is the time interval for which the IO has to be stopped before doing snapshot operation
	IoWaitTime = 10
	//TotalWaitTime is the max time duration to wait for doing snapshot operation on all the replicas
	TotalWaitTime = 60
)

// CStorVolume a wrapper for CStorVolume object
type CStorVolume struct {
	// actual cstorvolume object
	object *apis.CStorVolume
}

// CStorVolumeList is a list of cstorvolume objects
type CStorVolumeList struct {
	// list of cstor volumes
	items []*CStorVolume
}

// ListBuilder enables building
// an instance of CstorVolumeList
type ListBuilder struct {
	list    *CStorVolumeList
	filters PredicateList
}

//CVReplicationDetails enables to update RF,CF and
//known replicas into etcd
type CVReplicationDetails struct {
	VolumeName        string `json:"volumeName"`
	ReplicationFactor int    `json:"replicationFactor"`
	ConsistencyFactor int    `json:"consistencyFactor"`
	ReplicaID         string `json:"replicaId"`
	ReplicaGUID       uint64 `json:"replicaZvolGuid,string"`
}

//CStorVolumeConfig embed CVReplicationDetails and Kubeclient of
//corresponding namespace
type CStorVolumeConfig struct {
	*CVReplicationDetails
	*Kubeclient
}

// Conditions enables building CRUD operations on cstorvolume conditions
type Conditions []apis.CStorVolumeCondition

// GetResizeCondition will return resize condtion related to
// cstorvolume condtions
func GetResizeCondition() apis.CStorVolumeCondition {
	resizeConditions := apis.CStorVolumeCondition{
		Type:               apis.CStorVolumeResizing,
		Status:             apis.ConditionInProgress,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "Resizing",
		Message:            "Triggered resize by changing capacity in spec",
	}
	return resizeConditions
}

// NewListBuilder returns a new instance
// of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CStorVolumeList{}}
}

// WithAPIList builds the list of cstorvolume
// instances based on the provided
// cstorvolume api instances
func (b *ListBuilder) WithAPIList(list *apis.CStorVolumeList) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		c := c
		b.list.items = append(b.list.items, &CStorVolume{object: &c})
	}
	return b
}

// List returns the list of cstorvolume (cv)
// instances that was built by this
// builder
func (b *ListBuilder) List() *CStorVolumeList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &CStorVolumeList{}
	for _, cv := range b.list.items {
		if b.filters.all(cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CStorVolumeList
func (l *CStorVolumeList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cstorvolume instance
type Predicate func(*CStorVolume) bool

// IsHealthy returns true if the CVR is in
// healthy state
func (c *CStorVolume) IsHealthy() bool {
	return c.object.Status.Phase == "Healthy"
}

// IsHealthy is a predicate to filter out cstorvolumes
// which is healthy
func IsHealthy() Predicate {
	return func(c *CStorVolume) bool {
		return c.IsHealthy()
	}
}

// IsResizePending return true if resize is in progress
func (c *CStorVolume) IsResizePending() bool {
	curCapacity := c.object.Status.Capacity
	desiredCapacity := c.object.Spec.Capacity
	// Cmp returns 0 if the curCapacity is equal to desiredCapacity,
	// -1 if the curCapacity is less than desiredCapacity, or 1 if the
	// curCapacity is greater than desiredCapacity.
	return curCapacity.Cmp(desiredCapacity) == -1
}

// IsDRFPending return true if drf update is required else false
// Steps to verify whether drf is required
// 1. Read DesiredReplicationFactor configurations from istgt conf file
// 2. Compare the value with spec.DesiredReplicationFactor and return result
func (c *CStorVolume) IsDRFPending() bool {
	fileOperator := util.RealFileOperator{}
	ConfFileMutex.Lock()
	//If it has proper config then we will get --> "  DesiredReplicationFactor 3"
	i, gotConfig, err := fileOperator.GetLineDetails(IstgtConfPath, DesiredReplicationFactorKey)
	ConfFileMutex.Unlock()
	if err != nil || i == -1 {
		klog.Infof("failed to get %s config details error: %v",
			DesiredReplicationFactorKey,
			err,
		)
		return false
	}
	// gotConfig will have "  DesiredReplicationFactor  3" and we will extract
	// numeric character from output
	valueStr := strings.TrimFunc(gotConfig, func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		klog.Infof("failed to parse %s error: %v", valueStr, err)
		return false
	}
	return value == c.object.Spec.DesiredReplicationFactor
}

// GetCVCondition returns corresponding cstorvolume condition based argument passed
func (c *CStorVolume) GetCVCondition(
	condType apis.CStorVolumeConditionType) apis.CStorVolumeCondition {
	for _, cond := range c.object.Status.Conditions {
		if condType == cond.Type {
			return cond
		}
	}
	return apis.CStorVolumeCondition{}
}

// IsConditionPresent returns true if condition is available
func (c *CStorVolume) IsConditionPresent(condType apis.CStorVolumeConditionType) bool {
	for _, cond := range c.object.Status.Conditions {
		if condType == cond.Type {
			return true
		}
	}
	return false
}

// PredicateList holds a list of cstor volume
// based predicates
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided cstorvolume
// instance
func (l PredicateList) all(c *CStorVolume) bool {
	for _, check := range l {
		if !check(c) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cstorvolume has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// NewForAPIObject returns a new instance of cstorvolume
func NewForAPIObject(obj *apis.CStorVolume) *CStorVolume {
	return &CStorVolume{
		object: obj,
	}
}

// AddCondition appends the new condition to existing conditions
func (c Conditions) AddCondition(cond apis.CStorVolumeCondition) []apis.CStorVolumeCondition {
	c = append(c, cond)
	return c
}

// DeleteCondition deletes the condition from conditions
func (c Conditions) DeleteCondition(cond apis.CStorVolumeCondition) []apis.CStorVolumeCondition {
	newConditions := []apis.CStorVolumeCondition{}
	for _, condObj := range c {
		if condObj.Type != cond.Type {
			newConditions = append(newConditions, condObj)
		}
	}
	return newConditions
}

// UpdateCondition updates the condition if it is present in Conditions
func (c Conditions) UpdateCondition(cond apis.CStorVolumeCondition) []apis.CStorVolumeCondition {
	for i, condObj := range c {
		if condObj.Type == cond.Type {
			c[i] = cond
		}
	}
	return c
}

// BuildConfigData builds data based on the CVReplicationDetails
func (csr *CVReplicationDetails) BuildConfigData() map[string]string {
	data := map[string]string{}
	// Since we know what to update in istgt.conf file so constructing
	// key and value pairs
	// key represents what kind of configurations
	// value represents corresponding value for that key
	// TODO: Improve below code by exploring different options
	key := fmt.Sprintf("  ReplicationFactor")
	value := fmt.Sprintf("  ReplicationFactor %d", csr.ReplicationFactor)
	data[key] = value
	key = fmt.Sprintf("  ConsistencyFactor")
	value = fmt.Sprintf("  ConsistencyFactor %d", csr.ConsistencyFactor)
	data[key] = value
	key = fmt.Sprintf("  Replica %s", csr.ReplicaID)
	value = fmt.Sprintf("  Replica %s %d", csr.ReplicaID, csr.ReplicaGUID)
	data[key] = value
	return data
}

// UpdateConfig updates target configuration file by building data
func (csr *CVReplicationDetails) UpdateConfig() error {
	configData := csr.BuildConfigData()
	fileOperator := util.RealFileOperator{}
	ConfFileMutex.Lock()
	err := fileOperator.UpdateOrAppendMultipleLines(IstgtConfPath, configData, 0644)
	ConfFileMutex.Unlock()
	return err
}

// Validate verifies whether CStorReplication data read on wire is valid or not
func (csr *CVReplicationDetails) Validate() error {
	if csr.VolumeName == "" {
		return errors.Errorf("volume name can not be empty")
	}
	if csr.ReplicaID == "" {
		return errors.Errorf("replicaKey can not be empty to perform "+
			"volume %s update", csr.VolumeName)
	}
	if csr.ReplicaGUID == 0 {
		return errors.Errorf("replicaKey can not be empty to perform "+
			"volume %s update", csr.VolumeName)
	}
	if csr.ReplicationFactor == 0 {
		return errors.Errorf("replication factor can't be %d",
			csr.ReplicationFactor)
	}
	if csr.ConsistencyFactor == 0 {
		return errors.Errorf("consistencyFactor factor can't be %d",
			csr.ReplicationFactor)
	}
	return nil
}

// UpdateCVWithReplicationDetails updates the cstorvolume with known replicas
// and updated replication details
func (csc *CStorVolumeConfig) UpdateCVWithReplicationDetails() error {
	err := csc.Validate()
	if err != nil {
		return errors.Wrapf(err, "validate errors")
	}
	cv, err := csc.Get(csc.VolumeName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cstorvolume")
	}
	if len(cv.Status.ReplicaDetails.KnownReplicas) >= cv.Spec.DesiredReplicationFactor {
		return errors.Errorf("can not update cstorvolume %s known replica"+
			" count %d is greater than or equal to desired replication factor %d",
			cv.Name, len(cv.Status.ReplicaDetails.KnownReplicas),
			cv.Spec.DesiredReplicationFactor,
		)
	}
	if cv.Spec.ReplicationFactor > csc.ReplicationFactor {
		return errors.Errorf("requested replication factor {%d}"+
			" can not be smaller than existing replication factor {%d}",
			csc.ReplicationFactor, cv.Spec.ReplicationFactor,
		)
	}
	if cv.Spec.ConsistencyFactor > csc.ConsistencyFactor {
		return errors.Errorf("requested consistencyFactor factor {%d}"+
			" can not be smaller than existing consistencyFactor factor {%d}",
			csc.ReplicationFactor, cv.Spec.ConsistencyFactor,
		)
	}
	cv.Spec.ReplicationFactor = csc.ReplicationFactor
	cv.Spec.ConsistencyFactor = csc.ConsistencyFactor
	if cv.Status.ReplicaDetails.KnownReplicas == nil {
		cv.Status.ReplicaDetails.KnownReplicas = map[string]uint64{}
	}
	cv.Status.ReplicaDetails.KnownReplicas[csc.ReplicaID] = csc.ReplicaGUID
	_, err = csc.Update(cv)
	if err != nil {
		err = csc.UpdateConfig()
	}
	return err
}

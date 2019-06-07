/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	ndmclientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
)

//TODO: Update the file with latest pattern
const (
	// StorageNodePredicateKey is the key for StorageNodePredicate function.
	FilterInactive       = "filterInactive"
	FilterNonInactive    = "filterNonInactive"
	FilterClaimedDevices = "filterClaimedDevices"
	InActiveStatus       = "Inactive"
)

// DefaultDiskCount is a map containing the default disk count of various raid types.
var DefaultDiskCount = map[string]int{
	string(apis.PoolTypeMirroredCPV): int(apis.MirroredBlockDeviceCountCPV),
	string(apis.PoolTypeStripedCPV):  int(apis.StripedBlockDeviceCountCPV),
	string(apis.PoolTypeRaidzCPV):    int(apis.RaidzBlockDeviceCountCPV),
	string(apis.PoolTypeRaidz2CPV):   int(apis.Raidz2BlockDeviceCountCPV),
}

// KubernetesClient is the kubernetes client which will implement block device actions/behaviours
type KubernetesClient struct {
	// Kubeclientset is a standard kubernetes clientset
	Kubeclientset kubernetes.Interface

	// Clientset is a ndm custom resource package generated for ndm custom API group
	Clientset ndmclientset.Interface

	//Namespace is namespace where blockdevice is available
	Namespace string
}

type errs []error

// SpcObjectClient is the kubernetes client perform block devie operations in
// case of manual provisioning
type SpcObjectClient struct {
	*KubernetesClient
	Spc *apis.StoragePoolClaim
}

// BlockDevice is a wrapper over BlockDevice api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers
type BlockDevice struct {
	*ndm.BlockDevice
	errs
}

// BlockDeviceList is a wrapper over BlockDeviceList api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers
type BlockDeviceList struct {
	*ndm.BlockDeviceList
	errs
}

// BuildOptionFunc is the typed function to build BlockDevice object.
type BuildOptionFunc func(*BlockDevice)

// predicate is the typed predicate function to validate BlockDevice object.
type predicate func(*BlockDevice) (message string, ok bool)

// filterOptionFunc is the typed function to filter BlockDevice objects.
type filterOptionFunc func(original *BlockDeviceList) *BlockDeviceList

// BlockDeviceInterface abstracts operations on BlockDevice entity.
// Different orchestrators may need to implement this interface.
type BlockDeviceInterface interface {
	Get(name string, opts metav1.GetOptions) (*BlockDevice, error)
	List(opts metav1.ListOptions) (*BlockDeviceList, error)
	Create(*ndm.BlockDevice) (*BlockDevice, error)
}

// checkPredicatesFuncs is an array of check predicate functions.
var checkPredicatesFuncs = [...]predicate{
	checkName,
}

// filterPredicatesFuncMap is an array of filter predicate functions
// filter predicates should be tunable by client.
var filterOptionFuncMap = map[string]filterOptionFunc{
	FilterInactive:       filterInactive,
	FilterNonInactive:    filterNonInactive,
	FilterClaimedDevices: filterClaimedDevices,
}

// predicateFailedError returns the predicate error which is provided to this function as an argument
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// New is a constructor returns a new instance of block device
func New(opts ...BuildOptionFunc) *BlockDevice {
	r := &BlockDevice{BlockDevice: &ndm.BlockDevice{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Build returns the final instance of BlockDevice
func (bd *BlockDevice) Build() (*ndm.BlockDevice, []error) {
	bd.validate()
	if len(bd.errs) > 0 {
		return nil, bd.errs
	}
	return bd.BlockDevice, nil
}

// validate validates the block device object against the check predicates.
func (bd *BlockDevice) validate() {
	for _, c := range checkPredicatesFuncs {
		msg, ok := c(bd)
		if !ok {
			bd.errs = append(bd.errs, predicateFailedError(msg))
		}
	}
}

// WithName method fills the name field of BlockDevice object.
func (bd *BlockDevice) WithName(name string) *BlockDevice {
	WithName(name)(bd)
	return bd
}

// WithName function is used by WithName method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithName(name string) BuildOptionFunc {
	return func(bd *BlockDevice) {
		bd.BlockDevice.Name = name
	}
}

// WithState method fills the name field of BlockDevice object.
func (bd *BlockDevice) WithState(state string) *BlockDevice {
	WithState(state)(bd)
	return bd
}

// WithState function is used by WithState method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithState(state string) BuildOptionFunc {
	return func(bd *BlockDevice) {
		bd.BlockDevice.Status.State = state
	}
}

//checkName validate the name field of BlockDevice object.
func checkName(db *BlockDevice) (string, bool) {
	if db.BlockDevice.Name == "" {
		//TODO: Think about having some good organization in putting error messages.
		return "blockDevice name field on the object may not be empty", false
	}
	return "", true
}

// Filter adds filters on which the blockdevice has to be filtered
func (bdl *BlockDeviceList) Filter(predicateKeys ...string) *BlockDeviceList {
	// Initialize filtered block device list
	filteredBlockDeviceList := &BlockDeviceList{
		BlockDeviceList: &ndm.BlockDeviceList{},
		errs:            nil,
	}
	errMsg, ok := bdl.Hasitems()
	if !ok {
		filteredBlockDeviceList.errs = append(filteredBlockDeviceList.errs, errors.New(errMsg))
		return filteredBlockDeviceList
	}
	filteredBlockDeviceList = bdl
	for _, key := range predicateKeys {
		filteredBlockDeviceList = filterOptionFuncMap[key](filteredBlockDeviceList)
	}
	return filteredBlockDeviceList
}

//filterInactive filter and give out all the inactive block device
func filterInactive(orignialList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &ndm.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range orignialList.Items {
		if device.Status.State == InActiveStatus {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

//filterNonInactive give out all the block device except inactive block devices
func filterNonInactive(orignialList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &ndm.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range orignialList.Items {
		if !(device.Status.State == InActiveStatus) {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

func filterClaimedDevices(orignialList *BlockDeviceList) *BlockDeviceList {
	filteredList := &BlockDeviceList{
		BlockDeviceList: &ndm.BlockDeviceList{},
		errs:            nil,
	}
	for _, device := range orignialList.Items {
		if device.Status.ClaimState == ndm.BlockDeviceClaimed {
			filteredList.Items = append(filteredList.Items, device)
		}
	}
	return filteredList
}

// Hasitems checks whether the BlockDeviceList contains BlockDevices
func (bdl *BlockDeviceList) Hasitems() (string, bool) {
	if bdl == nil || bdl.BlockDeviceList == nil || bdl.Items == nil {
		return "No item found in blockdevice list", false
	}
	return "", true
}

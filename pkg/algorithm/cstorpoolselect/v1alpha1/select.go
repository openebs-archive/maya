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
	"errors"
	"strings"
	"text/template"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstorpool/v1alpha2"
	cvr "github.com/openebs/maya/pkg/cstorvolumereplica/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// cvrListFn abstracts fetching of a list of cstor
// volume replicas
type cvrListFn func(namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error)

type labelKey string

const (
	// preferReplicaAntiAffinty is the label key
	// that refers to preferring of replica
	// anti affinity policy
	preferReplicaAntiAffinityLabel labelKey = "openebs.io/preferred-replica-anti-affinity"

	// replicaAntiAffinty is the label key
	// that refers to replica anti affinity policy
	replicaAntiAffinityLabel labelKey = "openebs.io/replica-anti-affinity"
)

type annotationKey string

const (
	// scheduleOnHost is the annotation key
	// that refers to hostname to schedule
	// the replica
	scheduleOnHostAnnotation annotationKey = "volume.kubernetes.io/selected-node"
)

type priority int

const (
	// lowPriority refers to the priority
	// given to a selection policy
	lowPriority priority = 1

	// mediumPriority refers to the priority
	// given to a selection policy
	mediumPriority priority = 2

	// highPriority refers to the priority
	// given to a selection policy
	highPriority priority = 3
)

// policyName is a type that caters to
// naming of various pool selection
// policies
type policyName string

const (
	// antiAffinityLabelPolicy is the name of the
	// policy that applies anti-affinity rule against
	// storage placement
	antiAffinityLabelPolicy policyName = "anti-affinity-label"

	// preferAntiAffinityLabelPolicy is the name of
	// the policy that does a best effort while applying
	// anti-affinity rule against storage placement
	preferAntiAffinityLabelPolicy policyName = "prefer-anti-affinity-label"

	// scheduleOnHostAnnotationPolicy is the name of
	// the policy that selects the given host to
	// place storage
	scheduleOnHostAnnotationPolicy policyName = "schedule-on-host"

	// preferScheduleonHostPolicy is the name of
	// the policy that does a best effort to select
	// the given host to place storage
	preferScheduleOnHostAnnotationPolicy policyName = "prefer-schedule-on-host"
)

// policy exposes contracts that need
// to be satisfied by any pool selection
// implementation
type policy interface {
	priority() priority
	name() policyName
	filter(*csp.CSPList) (*csp.CSPList, error)
}

// scheduleOnHost is a pool selection
// implementation
type scheduleOnHost struct {
	// hostName holds the name of the
	// host on which storage needs to
	// be scheduled
	hostName string
}

// priority returns the priority of the
// policy implementation
func (p scheduleOnHost) priority() priority {
	return mediumPriority
}

// name returns the name of the policy
// implementation
func (p scheduleOnHost) name() policyName {
	return scheduleOnHostAnnotationPolicy
}

// filter selects the pools available on the host
// for which the policy has been applied
func (p scheduleOnHost) filter(pools *csp.CSPList) (*csp.CSPList, error) {
	if p.hostName == "" {
		return pools, nil
	}
	filteredPools := pools.Filter(csp.HasAnnotation(string(scheduleOnHostAnnotation), p.hostName))
	return filteredPools, nil
}

// preferScheduleOnHost is pool selection
// implementation
type preferScheduleOnHost struct {
	scheduleOnHost
}

// priority return the priority of the policy
// implementation
func (p preferScheduleOnHost) priority() priority {
	return mediumPriority
}

// name returns the name of the policy
// implementation
func (p preferScheduleOnHost) name() policyName {
	return preferScheduleOnHostAnnotationPolicy
}

// filter piggybacks on scheduleOnHost policy with
// the difference being this logic returns the
// provided pools if no pools are found on the host
func (p preferScheduleOnHost) filter(pools *csp.CSPList) (*csp.CSPList, error) {
	plist, err := p.scheduleOnHost.filter(pools)
	if err != nil {
		return nil, err
	}
	if len(plist.GetPoolUIDs()) == 0 {
		return pools, nil
	}
	return plist, nil
}

// antiAffinityLabel is a pool selection
// policy implementation
type antiAffinityLabel struct {
	labelSelector string

	// cvrList holds the function to list
	// cstor volume replica which is useful
	// mocking
	cvrList cvrListFn
}

func defaultCVRList() cvrListFn {
	return func(namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error) {
		return cvr.NewKubeclient(cvr.WithNamespace(namespace)).List(opts)
	}
}

// priority returns the priority of
// this policy
func (p antiAffinityLabel) priority() priority {
	return lowPriority
}

// name returns the name of this
// policy
func (p antiAffinityLabel) name() policyName {
	return antiAffinityLabelPolicy
}

// filter excludes the pool(s) if they are
// already associated with the label
// selector. In other words, it applies anti
// affinity rule against the provided list of
// pools.
func (p antiAffinityLabel) filter(pools *csp.CSPList) (*csp.CSPList, error) {
	if p.labelSelector == "" {
		return pools, nil
	}
	// pools that are already associated with
	// this label should be excluded
	//
	// NOTE: we try without giving any namespace
	// so that it lists from all available
	// namespaces
	cvrs, err := p.cvrList("", metav1.ListOptions{LabelSelector: p.labelSelector})
	if err != nil {
		return nil, err
	}

	exclude := cvr.ListBuilder().WithAPIList(cvrs).List().GetPoolUIDs()
	return pools.Filter(csp.IsNotUID(exclude...)), nil
}

// preferAntiAffinityLabel is a pool
// selection policy implementation
type preferAntiAffinityLabel struct {
	antiAffinityLabel
}

// name returns the name of this policy
func (p preferAntiAffinityLabel) name() policyName {
	return preferAntiAffinityLabelPolicy
}

// filter piggybacks on antiAffinityLabel policy
// with the difference being; this logic returns all
// the provided pools if there are no pools that
// satisfy antiAffinity rule
func (p preferAntiAffinityLabel) filter(pools *csp.CSPList) (*csp.CSPList, error) {
	plist, err := p.antiAffinityLabel.filter(pools)
	if err != nil {
		return plist, err
	}
	if len(plist.GetPoolUIDs()) > 0 {
		return plist, nil
	}
	return pools, nil
}

type executionMode string

const (
	// multiExection enables execution of
	// more than one policy during a selection
	multiExecution executionMode = "multi-mode"

	// singleExecution enables execution of
	// only one policy during a seclection
	singleExection executionMode = "single-mode"
)

type policyList struct {
	items map[priority][]policy
}

func (pl *policyList) getAll() []policy {
	if len(pl.items) == 0 {
		return nil
	}
	var all []policy
	for _, policies := range pl.items {
		all = append(all, policies...)
	}
	return all
}

func (pl *policyList) add(p policy) {
	pl.items[p.priority()] = append(pl.items[p.priority()], p)
}

func (pl *policyList) sortByPriority() []policy {
	var sorted []policy
	if len(pl.items) == 0 {
		return sorted
	}
	for i := highPriority; i >= lowPriority; i-- {
		if len(pl.items[i]) == 0 {
			continue
		}
		sorted = append(sorted, pl.items[i]...)
	}
	return sorted
}

type policyListPredicate func(*policyList) bool

func hasPolicy(name policyName) policyListPredicate {
	return func(pl *policyList) bool {
		if len(pl.items) == 0 {
			return false
		}
		all := pl.getAll()
		for _, policy := range all {
			if policy.name() == name {
				return true
			}
		}
		return false
	}
}

func hasHighPriorityPolicy() policyListPredicate {
	return func(pl *policyList) bool {
		if len(pl.items) == 0 {
			return false
		}
		return len(pl.items[highPriority]) != 0
	}
}

func hasMediumPriorityPolicy() policyListPredicate {
	return func(pl *policyList) bool {
		if len(pl.items) == 0 {
			return false
		}
		return len(pl.items[mediumPriority]) != 0
	}
}

func hasLowPriorityPolicy() policyListPredicate {
	return func(pl *policyList) bool {
		if len(pl.items) == 0 {
			return false
		}
		return len(pl.items[lowPriority]) != 0
	}
}

func (pl *policyList) getTopPriority() policy {
	if hasHighPriorityPolicy()(pl) {
		return pl.items[highPriority][0]
	} else if hasMediumPriorityPolicy()(pl) {
		return pl.items[mediumPriority][0]
	} else if hasLowPriorityPolicy()(pl) {
		return pl.items[lowPriority][0]
	}
	return nil
}

// selection enables selecting required pools
// based on the registered policies
//
// NOTE:
//  There can be cases where multiple policies
// can be set to determine the required pools
//
// NOTE:
//  This code will evolve as we try implementing
// different set of policies
type selection struct {
	// list of original pools aginst whom
	// selection will be made
	pools *csp.CSPList

	// selection is based on these policies
	policies *policyList

	// mode flags if selection can consider
	// multiple policies to select the pools
	mode executionMode
}

// buildOption is a typed function that
// abstracts configuring a selection instance
type buildOption func(*selection)

func withDefaultSelection(s *selection) {
	if string(s.mode) == "" {
		s.mode = singleExection
	}
}

// newSelection returns a new instance of
// selection
func newSelection(pools *csp.CSPList, opts ...buildOption) *selection {
	s := &selection{pools: pools, policies: &policyList{map[priority][]policy{}}}
	for _, o := range opts {
		if o != nil {
			o(s)
		}
	}
	withDefaultSelection(s)
	return s
}

// hasPolicy determines if the provided policy
// is part of the selection
func (s *selection) hasPolicy(p policyName) bool {
	return hasPolicy(p)(s.policies)
}

// hasPreferAntiAffinityLabel determines if
// prefer anti affinity label is part of
// the selection
func (s *selection) hasPreferAntiAffinityLabel() bool {
	return s.hasPolicy(preferAntiAffinityLabelPolicy)
}

// hasAntiAffinityLabel determines if anti affinity
// label is part of the selection
func (s *selection) hasAntiAffinityLabel() bool {
	return s.hasPolicy(antiAffinityLabelPolicy)
}

// ExecutionMode sets the execution mode
// against the provided selection instance
func ExecutionMode(m executionMode) buildOption {
	return func(s *selection) {
		s.mode = m
	}
}

// PreferAntiAffinityLabel adds anti affinity label
// as a preferred policy to be used during pool
// selection
func PreferAntiAffinityLabel(lbl string) buildOption {
	return func(s *selection) {
		p := preferAntiAffinityLabel{antiAffinityLabel{labelSelector: lbl, cvrList: defaultCVRList()}}
		s.policies.add(p)
	}
}

// AntiAffinityLabel adds anti affinity label
// as a policy to be used during pool selection
func AntiAffinityLabel(lbl string) buildOption {
	return func(s *selection) {
		p := antiAffinityLabel{labelSelector: lbl, cvrList: defaultCVRList()}
		s.policies.add(p)
	}
}

// PreferScheduleOnHostAnnotation adds preferScheduleOnHost
// as a policy to be used during pool selection
func PreferScheduleOnHostAnnotation(hostNameAnnotation string) buildOption {
	hostName := strings.TrimPrefix(hostNameAnnotation, string(scheduleOnHostAnnotation)+"=")
	return func(s *selection) {
		p := preferScheduleOnHost{scheduleOnHost{hostName: hostName}}
		s.policies.add(p)
	}
}

// GetPolicies returns the appropriate selection
// policies based on the provided values
func GetPolicies(values ...string) []buildOption {
	var opts []buildOption
	for _, val := range values {
		if strings.Contains(val, string(scheduleOnHostAnnotation)) {
			opts = append(opts, PreferScheduleOnHostAnnotation(val))
		} else if strings.Contains(val, string(preferReplicaAntiAffinityLabel)) {
			opts = append(opts, PreferAntiAffinityLabel(val))
		} else if strings.Contains(val, string(replicaAntiAffinityLabel)) {
			opts = append(opts, AntiAffinityLabel(val))
		}
	}
	return opts
}

// validate runs some validations/checks
// against this selection instance
func (s *selection) validate() error {
	if s.hasAntiAffinityLabel() && s.hasPreferAntiAffinityLabel() {
		return errors.New("invalid selection: both antiAffinityLabel and preferAntiAffinityLabel policies can not be together")
	}
	return nil
}

// filter returns the final list of pools that
// gets selected, after passing the original list
// of pools through the registered selection policies
func (s *selection) filter() (*csp.CSPList, error) {
	var err error
	if s.policies == nil || len(s.policies.items) == 0 || s.pools == nil || len(s.pools.GetPoolUIDs()) == 0 {
		return s.pools, nil
	}
	// make a copy of original pool UIDs
	filtered := csp.ListBuilder().WithList(s.pools).List()
	// Sorting policies based on the priority
	policies := s.policies.sortByPriority()
	// Executing policy filters
	for _, policy := range policies {
		filtered, err = policy.filter(filtered)
		if err != nil {
			return nil, err
		}
		// stopping here if running as
		// singleExecution mode
		if s.mode == singleExection {
			break
		}
	}
	return filtered, nil
}

// Filter will filter the provided pools
// based on pool selection policies
func Filter(entries *csp.CSPList, opts ...buildOption) (*csp.CSPList, error) {
	if entries == nil {
		return entries, nil
	}
	s := newSelection(entries, opts...)
	err := s.validate()
	if err != nil {
		return nil, err
	}
	return s.filter()
}

// FilterPoolIDs will filter the provided pools
// based on pool selection policies
func FilterPoolIDs(entries *csp.CSPList, opts []buildOption) ([]string, error) {
	plist, err := Filter(entries, opts...)
	if err != nil {
		return nil, err
	}
	return plist.GetPoolUIDs(), nil
}

// TemplateFunctions exposes a few functions as
// go template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"cspGetPolicies":        GetPolicies,
		"cspFilterPoolIDs":      FilterPoolIDs,
		"cspAntiAffinity":       AntiAffinityLabel,
		"cspPreferAntiAffinity": PreferAntiAffinityLabel,
		"preferScheduleOnHost":  PreferScheduleOnHostAnnotation,
	}
}

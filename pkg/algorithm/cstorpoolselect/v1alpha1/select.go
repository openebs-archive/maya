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

type labelKey string

const (
	preferReplicaAntiAffinityLabel labelKey = "openebs.io/preferred-replica-anti-affinity"
	replicaAntiAffinityLabel       labelKey = "openebs.io/replica-anti-affinity"
)

// cvrListFn abstracts fetching of a list of cstor
//  volume replicas
type cvrListFn func(namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error)

// policyName is a type that caters to
// naming of various pool selection
// policies
type policyName string

const (
	// antiAffinityLabelPolicyName is the name of the
	// policy that applies anti-affinity rule based on
	// label
	antiAffinityLabelPolicyName policyName = "anti-affinity-label"

	// preferAntiAffinityLabelPolicyName is the name of
	// the policy that does a best effort while applying
	// anti-affinity rule based on label
	preferAntiAffinityLabelPolicyName policyName = "prefer-anti-affinity-label"
)

// policy exposes the contracts that need
// to be satisfied by any pool selection
// implementation
type policy interface {
	name() policyName
	filter([]string) ([]string, error)
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

// defaultCVRList is the default
// implementation of cvrListFn
func defaultCVRList() cvrListFn {
	return cvr.KubeClient().List
}

// name returns the name of this
// policy
func (p antiAffinityLabel) name() policyName {
	return antiAffinityLabelPolicyName
}

// filter excludes the pool(s) if they are
// already associated with the label
// selector. In other words, it applies anti
// affinity rule against the provided list of
// pools.
func (l antiAffinityLabel) filter(poolUIDs []string) ([]string, error) {
	if l.labelSelector == "" {
		return poolUIDs, nil
	}
	// pools that are already associated with
	// this label should be excluded
	//
	// NOTE: we try without giving any namespace
	// so that it lists from all available
	// namespaces
	cvrs, err := l.cvrList("", metav1.ListOptions{LabelSelector: l.labelSelector})
	if err != nil {
		return nil, err
	}
	exclude := cvr.ListBuilder().WithListObject(cvrs).List().GetPoolUIDs()
	plist := csp.ListBuilder().WithUIDs(poolUIDs...).List()
	return plist.FilterUIDs(csp.IsNotUID(exclude...)), nil
}

// preferAntiAffinityLabel is a pool
// selection policy implementation
type preferAntiAffinityLabel struct {
	antiAffinityLabel
}

// name returns the name of this policy
func (p preferAntiAffinityLabel) name() policyName {
	return preferAntiAffinityLabelPolicyName
}

// filter piggybacks on antiAffinityLabel policy
// with the difference being; this logic returns all
// the provided pools if there are no pools that
// satisfy antiAffinity rule
func (p preferAntiAffinityLabel) filter(poolUIDs []string) ([]string, error) {
	plist, err := p.antiAffinityLabel.filter(poolUIDs)
	if err != nil {
		return nil, err
	}
	if len(plist) > 0 {
		return plist, nil
	}
	return poolUIDs, nil
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
	poolUIDs []string

	// selection is based on these policies
	policies []policy
}

// buildOption is a typed function that
// abstracts configuring a selection instance
type buildOption func(*selection)

// newSelection returns a new instance of
// selection
func newSelection(poolUIDs []string, opts ...buildOption) *selection {
	s := &selection{poolUIDs: poolUIDs}
	for _, o := range opts {
		if o != nil {
			o(s)
		}
	}
	return s
}

// isPolicy determines if the provided policy
// needs to be considered during selection
func (s *selection) isPolicy(p policyName) bool {
	for _, pol := range s.policies {
		if pol.name() == p {
			return true
		}
	}
	return false
}

// isPreferAntiAffinityLabel determines if
// prefer anti affinity label needs to be
// considered during selection
func (s *selection) isPreferAntiAffinityLabel() bool {
	return s.isPolicy(preferAntiAffinityLabelPolicyName)
}

// isAntiAffinityLabel determines if anti affinity
// label needs to be considered during
// selection
func (s *selection) isAntiAffinityLabel() bool {
	return s.isPolicy(antiAffinityLabelPolicyName)
}

// PreferAntiAffinityLabel adds anti affinity label
// as a preferred policy to be used during pool
// selection
func PreferAntiAffinityLabel(lbl string) buildOption {
	return func(s *selection) {
		p := preferAntiAffinityLabel{antiAffinityLabel{labelSelector: lbl, cvrList: defaultCVRList()}}
		s.policies = append(s.policies, p)
	}
}

// AntiAffinityLabel adds anti affinity label
// as a policy to be used during pool selection
func AntiAffinityLabel(lbl string) buildOption {
	return func(s *selection) {
		a := antiAffinityLabel{labelSelector: lbl, cvrList: defaultCVRList()}
		s.policies = append(s.policies, a)
	}
}

// GetBuildOptionByLabelSelector returns the appropriate
// buildOptions based on the input label
func GetBuildOptionByLabelSelector(labels ...string) []buildOption {
	var opts []buildOption
	for _, label := range labels {
		if strings.Contains(label, string(preferReplicaAntiAffinityLabel)) {
			opts = append(opts, PreferAntiAffinityLabel(label))
		} else if strings.Contains(label, string(replicaAntiAffinityLabel)) {
			opts = append(opts, AntiAffinityLabel(label))
		}
	}
	return opts
}

// validate runs some validations/checks
// against this selection instance
func (s *selection) validate() error {
	if s.isAntiAffinityLabel() && s.isPreferAntiAffinityLabel() {
		return errors.New("invalid selection: both antiAffinityLabel and preferAntiAffinityLabel policies can not be together")
	}
	return nil
}

// filter returns the final list of pools that
// gets selected, after passing the original list
// of pools through the registered selection policies
func (s *selection) filter() ([]string, error) {
	var (
		filtered []string
		err      error
	)
	if len(s.policies) == 0 {
		return s.poolUIDs, nil
	}
	// make a copy of original pool UIDs
	filtered = append(filtered, s.poolUIDs...)
	for _, policy := range s.policies {
		filtered, err = policy.filter(filtered)
		if err != nil {
			return nil, err
		}
	}
	return filtered, nil
}

// FilterWithBuildOptions will return filtered pool UIDs
// from the provided list based on pool
// selection options
func FilterWithBuildOptions(origPoolUIDs []string, opts []buildOption) ([]string, error) {
	return Filter(origPoolUIDs, opts...)
}

// Filter will return filtered pool UIDs
// from the provided list based on pool
// selection options
func Filter(origPoolUIDs []string, opts ...buildOption) ([]string, error) {
	if len(opts) == 0 {
		return origPoolUIDs, nil
	}
	s := newSelection(origPoolUIDs, opts...)
	err := s.validate()
	if err != nil {
		return nil, err
	}
	return s.filter()
}

// TemplateFunctions exposes a few functions as go template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"cspGetPolicyByLabelSelector": GetBuildOptionByLabelSelector,
		"cspFilter":                   FilterWithBuildOptions,
		"cspAntiAffinity":             AntiAffinityLabel,
		"cspPreferAntiAffinity":       PreferAntiAffinityLabel,
	}
}

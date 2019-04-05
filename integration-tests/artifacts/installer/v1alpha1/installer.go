package v1alpha1

import (
	"fmt"

	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha1"

	unstructuredUtil "github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Installer enables installing/uninstalling
// of resources in container orchestrator
type Installer struct {
	object     *unstructured.Unstructured
	predicates []Predicate
}

// Predicate abstracts the verification of
// a resource
type Predicate func(*unstructured.Unstructured) bool

// String returns the meta information of the
// installing resource
func (i *Installer) String() string {
	if i.object == nil {
		return "installer - object=nil"
	}
	return fmt.Sprintf("installer - name=%s namespace=%s kind=%s",
		i.object.GetName(),
		i.object.GetNamespace(),
		i.object.GetKind(),
	)
}

// GoString returns the go string of the installer
func (i *Installer) GoString() string {
	return i.String()
}

// Install triggers the installation of resource
// in the kubernetes cluster
func (i *Installer) Install() error {
	k, err := unstruct.KubeClient()
	if err != nil {
		return err
	}
	return k.Create(i.object)
}

// UnInstall triggers deletion of resource from
// kubernetes cluster
func (i *Installer) UnInstall() error {
	k, err := unstruct.KubeClient()
	if err != nil {
		return err
	}
	return k.Delete(i.object)
}

// Verify returns the installation of resource.
// It returns true if all the predicates passes
func (i *Installer) Verify() (bool, error) {
	k, err := unstruct.KubeClient()
	if err != nil {
		return false, err
	}
	obj, err := k.Get(
		i.object.GetName(),
		provider.WithGetNamespace(i.object.GetNamespace()),
		provider.WithGroupVersionKind(i.object.GroupVersionKind()),
	)
	if err != nil {
		return false, err
	}
	for _, p := range i.predicates {
		if p != nil {
			result := p(obj)
			if !result {
				return result, nil
			}
		}
	}
	return true, nil
}

// Builder abstracts the construction of the builder
type Builder struct {
	installer *Installer
	errs      []error
}

// BuilderForObject creates the installer builder from
// the object provided
func BuilderForObject(obj *unstructured.Unstructured) *Builder {
	b := &Builder{}
	if obj == nil {
		b.errs = append(b.errs, errors.Errorf("failed to build for object: nil unstruct instance provided"))
		return b
	}
	b.installer = &Installer{object: obj}
	return b
}

// BuilderForYaml creates the installer builder from the
// YAML provided
func BuilderForYaml(yaml string) *Builder {
	b := &Builder{}
	obj, err := unstruct.BuilderForYaml(yaml).Build()
	if err != nil {
		b.errs = append(b.errs, err)
		return b
	}
	b.installer = &Installer{object: obj.GetUnstructured()}
	return b
}

// AddCheck adds verify predicated for the installer
func (b *Builder) AddCheck(p ...Predicate) *Builder {
	for _, o := range p {
		if o != nil {
			b.installer.predicates = append(b.installer.predicates, o)
		}
	}
	return b
}

// Build triggers the building of the installer
func (b *Builder) Build() (*Installer, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%v", b.errs)
	}
	return b.installer, nil
}

// IsPodRunning returns true if the pod is in running
// state
func IsPodRunning() Predicate {
	return func(u *unstructured.Unstructured) bool {
		v := unstructuredUtil.GetNestedString(u.Object, "status", "phase")
		return v == "Running"
	}
}

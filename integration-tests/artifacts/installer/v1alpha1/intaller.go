package v1alpha1

import (
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// installer enables installation of a kubernetes
// resource
type installer struct {
	clientset client.Client
	namespace string
	object    *unstructured.Unstructured
	err       error
}

// buildOptions defines the abstraction to
// build the installer
type buildOptions func(i *installer)

// listBuildOptions defines the abstraction
// to build the list installer
type listBuildOptions func(i *listInstaller)

// Installer returns a new instance Installer
func Installer(bo ...buildOptions) *installer {
	i := &installer{}
	for _, b := range bo {
		if b != nil {
			b(i)
		}
	}
	return i
}

// FromObject instructs the installer to install
// the object from a unstructured instance
func (i *installer) FromObject(u *unstructured.Unstructured) *installer {
	i.object = u
	return i
}

// FromYAML instructs the installer to install
// the object from the given YAML
func (i *installer) FromYAML(doc string) *installer {
	obj, err := unstruct.BuilderForYaml(doc).Build()
	if err != nil {
		i.err = err
	}
	i.object = obj.GetUnstructured()
	return i
}

// Install triggers install process for the
// install client
func (i *installer) Install() error {
	if i.object != nil && i.namespace != "" {
		i.object.SetNamespace(i.namespace)
	}
	k, err := unstruct.KubeClient(unstruct.WithClient(i.clientset))
	if err != nil {
		return err
	}
	err = k.Create(i.object)
	return err
}

// Delete triggers delete process for the
// install client
func (i *installer) Delete() error {
	if i.object != nil && i.namespace != "" {
		i.object.SetNamespace(i.namespace)
	}
	k, err := unstruct.KubeClient(unstruct.WithClient(i.clientset))
	if err != nil {
		return err
	}
	err = k.Delete(i.object)
	return err
}

// listInstaller holds a list of install
// client
type listInstaller struct {
	errors     []error
	installers []*installer
	clientset  client.Client
	namespace  string
}

// ListInstaller returns a new instance of
// listInstaller
func ListInstaller(lbos ...listBuildOptions) *listInstaller {
	i := &listInstaller{}
	for _, lbo := range lbos {
		if lbo != nil {
			lbo(i)
		}
	}
	return i
}

// Fromobject instructs the list installer to install
// the given unstructured objects
func (l *listInstaller) FromObjects(objs ...*unstructured.Unstructured) *listInstaller {
	for _, obj := range objs {
		if obj != nil {
			i := Installer()
			if l.namespace != "" {
				i.namespace = l.namespace
			}
			if l.clientset != nil {
				i.clientset = l.clientset
			}
			l.installers = append(l.installers, i.FromObject(obj))
		}
	}
	return l
}

// FromYAMLs instructs the list installer to install
// the given YAML resources
func (l *listInstaller) FromYAMLs(doc string) *listInstaller {
	objs, err := unstruct.ListBuilderForYamls(doc).Build()
	if err != nil {
		l.errors = append(l.errors, err)
	}
	for _, unstructObj := range objs {
		if unstructObj != nil {
			obj := unstructObj.GetUnstructured()
			i := Installer()
			if l.namespace != "" {
				i.namespace = l.namespace
			}
			if l.clientset != nil {
				i.clientset = l.clientset
			}
			l.installers = append(l.installers, i.FromObject(obj))
		}
	}
	return l
}

// Install triggers the the installation process
// for the given list
func (l *listInstaller) Install() error {
	for _, i := range l.installers {
		err := i.Install()
		if err != nil {
			l.errors = append(l.errors, err)
		}
	}
	if len(l.errors) > 0 {
		return errors.Errorf("%v", l.errors)
	}
	return nil
}

// Delete triggers the the deletion process
// for the given list
func (l *listInstaller) Delete() error {
	for _, i := range l.installers {
		err := i.Delete()
		if err != nil {
			l.errors = append(l.errors, err)
		}
	}
	if len(l.errors) > 0 {
		return errors.Errorf("%v", l.errors)
	}
	return nil
}

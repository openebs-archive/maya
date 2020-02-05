package patch

import (
	"bytes"
	"fmt"

	"github.com/golang/glog"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Deployment ...
type Deployment struct {
	Object appsv1.Deployment
	Data   []byte
}

// NewDeployment ...
func NewDeployment() *Deployment {
	return &Deployment{}
}

// PreChecks ...
func (d *Deployment) PreChecks(from, to string) error {
	name := d.Object.Name
	if name == "" {
		return errors.Errorf("missing deployment name")
	}
	version := d.Object.Labels[OpenebsVersionLabel]
	if version != from && version != to {
		return errors.Errorf(
			"deployment version %s is neither %s nor %s",
			version,
			from,
			to,
		)
	}
	return nil
}

// Patch ...
func (d *Deployment) Patch(from, to string) error {
	buffer := bytes.Buffer{}
	client := deploy.NewKubeClient(deploy.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig"))
	version := d.Object.Labels[OpenebsVersionLabel]
	if version == to {
		glog.Infof("deployment already in %s version", to)
		return nil
	}
	if version == from {
		fmt.Println(string(d.Data))
		buffer.Reset()
		_, err := client.WithNamespace(d.Object.Namespace).Patch(
			d.Object.Name,
			types.StrategicMergePatchType,
			d.Data,
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to patch deployment %s",
				d.Object.Name,
			)
		}
		buffer.Reset()
		glog.Infof("deployment %s patched", d.Object.Name)
	}
	return nil
}

// Get ...
func (d *Deployment) Get(label, namespace string) error {
	client := deploy.NewKubeClient(deploy.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig"))
	deployments, err := client.WithNamespace(namespace).List(
		&metav1.ListOptions{
			LabelSelector: label,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to get deployment for %s", label)
	}
	if len(deployments.Items) != 1 {
		return errors.Errorf("no deployments found for label: %s  in namespace %s", label, namespace)
	}
	d.Object = deployments.Items[0]
	return nil
}

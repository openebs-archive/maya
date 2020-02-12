package patch

import (
	"time"

	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	retry "github.com/openebs/maya/pkg/util/retry"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// Deployment ...
type Deployment struct {
	Object *appsv1.Deployment
	Data   []byte
}

// NewDeployment ...
func NewDeployment() *Deployment {
	return &Deployment{}
}

// PreChecks ...
func (d *Deployment) PreChecks(from, to string) error {
	if d.Object == nil {
		return errors.Errorf("nil deployment object")
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
	klog.Info("patching deployment ", d.Object.Name)
	client := deploy.NewKubeClient(deploy.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig"))
	version := d.Object.Labels[OpenebsVersionLabel]
	if version == to {
		klog.Infof("deployment already in %s version", to)
		return nil
	}
	if version == from {
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
		err = retry.
			Times(60).
			Wait(5 * time.Second).
			Try(func(attempt uint) error {
				rolloutStatus, err1 := client.WithNamespace(d.Object.Namespace).
					RolloutStatus(d.Object.Name)
				if err1 != nil {
					return err1
				}
				if !rolloutStatus.IsRolledout {
					return errors.Errorf("failed to rollout: %s", rolloutStatus.Message)
				}
				return nil
			})
		if err != nil {
			return err
		}
		klog.Infof("deployment %s patched successfully", d.Object.Name)
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
	d.Object = &deployments.Items[0]
	return nil
}

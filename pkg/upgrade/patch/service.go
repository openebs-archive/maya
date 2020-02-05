package patch

import (
	"bytes"
	"fmt"

	"github.com/golang/glog"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Service ...
type Service struct {
	Object *corev1.Service
	Data   []byte
}

// NewService ...
func NewService() *Service {
	return &Service{}
}

// PreChecks ...
func (s *Service) PreChecks(from, to string) error {
	name := s.Object.Name
	if name == "" {
		return errors.Errorf("missing service name")
	}
	version := s.Object.Labels[OpenebsVersionLabel]
	if version != from && version != to {
		return errors.Errorf(
			"service version %s is neither %s nor %s",
			version,
			from,
			to,
		)
	}
	return nil
}

// Patch ...
func (s *Service) Patch(from, to string) error {
	buffer := bytes.Buffer{}
	client := svc.NewKubeClient(svc.WithKubeConfigPath("/home/user/.kube/config"))
	version := s.Object.Labels[OpenebsVersionLabel]
	if version == to {
		glog.Infof("service already in %s version", to)
		return nil
	}
	if version == from {
		patch := s.Data
		fmt.Println(patch)
		buffer.Reset()
		_, err := client.WithNamespace(s.Object.Namespace).Patch(
			s.Object.Name,
			types.StrategicMergePatchType,
			[]byte(patch),
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to patch service %s",
				s.Object.Name,
			)
		}
		buffer.Reset()
		glog.Infof("service %s patched", s.Object.Name)
	}
	return nil
}

// Get ...
func (s *Service) Get(label, namespace string) error {
	client := svc.NewKubeClient(svc.WithKubeConfigPath("/home/user/.kube/config"))
	service, err := client.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: label,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to get service for %s", label)
	}
	s.Object = &service.Items[0]
	return nil
}

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
package patch

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
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
	version := s.Object.Labels[string(apis.OpenEBSVersionKey)]
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
	klog.Info("patching service ", s.Object.Name)
	client := svc.NewKubeClient(svc.WithKubeConfigPath("/home/user/.kube/config"))
	version := s.Object.Labels[string(apis.OpenEBSVersionKey)]
	if version == to {
		klog.Infof("service already in %s version", to)
		return nil
	}
	if version == from {
		patch := s.Data
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
		klog.Infof("service %s patched", s.Object.Name)
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

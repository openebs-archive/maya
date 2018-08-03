/*
Copyright 2017 The OpenEBS Authors

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

package task

import (
	"fmt"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskSpecFetcher is the contract to fetch task
// specification that includes the task's meta specification
type TaskSpecFetcher interface {
	Fetch(taskName string) (runtask v1alpha1.RunTask, err error)
}

// K8sTaskSpecFetcher deals with fetching a task specifications
// from K8s cluster
//
// NOTE:
//  A task is a K8s ConfigMap
type K8sTaskSpecFetcher struct {
	// k8sClient to make K8s API calls
	//
	// NOTE:
	//  This is also helpful for mocking during UT
	k8sClient *m_k8s_client.K8sClient
}

// NewK8sTaskSpecFetcher returns a new instance of K8sTaskSpecFetcher based on the
// provided namespace.
//
// NOTE:
//  SearchNamespace refers to the K8s namespace where a task
// is expected to be found
func NewK8sTaskSpecFetcher(searchNamespace string) (*K8sTaskSpecFetcher, error) {
	kc, err := m_k8s_client.NewK8sClient(searchNamespace)
	if err != nil {
		return nil, err
	}

	return &K8sTaskSpecFetcher{
		k8sClient: kc,
	}, nil
}

// FetchSpec returns specifications of a provided task in yaml
// string format
//
// NOTE:
//  This is an implementation of TaskSpecFetcher interface
func (f *K8sTaskSpecFetcher) Fetch(taskName string) (runtask v1alpha1.RunTask, err error) {
	if len(taskName) == 0 {
		err = fmt.Errorf("failed to fetch runtask: nil task name was provided")
		return
	}

	cm, err := f.k8sClient.GetConfigMap(taskName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return
	}

	runtask = v1alpha1.RunTask{}
	runtask.Name = taskName
	runtask.Spec.Meta = cm.Data["meta"]
	runtask.Spec.Task = cm.Data["task"]
	runtask.Spec.PostRun = cm.Data["post"]

	return
}

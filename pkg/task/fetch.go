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

	"github.com/golang/glog"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskSpecFetcher is the contract to fetch task
// specification that includes the task's meta specification
type TaskSpecFetcher interface {
	Fetch(taskName string) (metaTaskYml string, taskYml string, err error)
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
func (f *K8sTaskSpecFetcher) Fetch(taskName string) (metaTaskYml string, taskYml string, err error) {
	if len(taskName) == 0 {
		return "", "", fmt.Errorf("Nil task name: Task can not be fetched")
	}

	cm, err := f.k8sClient.GetConfigMap(taskName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	// TODO
	// Validations if this CM is actually a Task

	// return the yaml string representation of metatask & task respectively
	metaTaskYml = cm.Data["meta"]
	if len(metaTaskYml) == 0 {
		return "", "", fmt.Errorf("Nil meta task specs: Fetched task is invalid '%s'", taskName)
	}

	taskYml = cm.Data["task"]
	if len(taskYml) == 0 {
		// This can be empty for get API calls
		glog.Warningf("Nil task specs: Will use meta task specs: Task: '%s' MetaTask: '%s'", taskName, metaTaskYml)
	}

	return metaTaskYml, taskYml, nil
}

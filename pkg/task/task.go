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

// instead of pkg/maya/template.go
package task

import (
	"fmt"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	m_k8s "github.com/openebs/maya/pkg/k8s"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// Task represents a task that is capable of being
// executed in a workflow. A task execution typically
// means invoking an API call e.g. a K8s API call.
type Task struct {
	// Values are the inputs that needs to be provided
	// to this task's template i.e. yaml
	values map[string]interface{}
	// yml represents the YAML representation of this task
	yml string
}

type taskExecutor struct {
	// identity is the id of the task
	identity string
	// objectName is the name of the object
	//
	// NOTE:
	//  object refers to the result of task execution
	objectName string
	// taskResultQueries is a set of queries that are run
	// against the object after successful execution of the task
	taskResultQueries []TaskResultQuery
	// taskPatch is a set of patches that get applied against
	// the task object
	taskPatch TaskPatch
	// metaTaskExec is the instance to be used to execute meta
	// operations on this task
	metaTaskExec *metaTaskExecutor
	// task has info related to the task
	task Task
	// k8sClient will be used to make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// NewTask returns a new instance of Task
func newTaskExecutor(identity, metaTaskYml, taskYml string, values map[string]interface{}) (*taskExecutor, error) {
	mte, err := newMetaTaskExecutor(identity, metaTaskYml, values)
	if err != nil {
		return nil, err
	}

	// client to make K8s API calls using the namespace on
	// which this task is supposed to be executed
	kc, err := m_k8s_client.NewK8sClient(mte.getRunNamespace())
	if err != nil {
		return nil, err
	}

	return &taskExecutor{
		identity:          identity,
		objectName:        mte.getObjectName(),
		taskResultQueries: mte.getTaskResultQueries(),
		taskPatch:         mte.getTaskPatch(),
		metaTaskExec:      mte,
		task: Task{
			yml:    taskYml,
			values: values,
		},
		k8sClient: kc,
	}, nil
}

// getMetaTaskExecutor gets the meta task executor value
func (m *taskExecutor) getMetaTaskExecutor() metaTaskExecutor {
	return *m.metaTaskExec
}

// Execute will execute the Task depending on informations
// available in Meta
func (m *taskExecutor) execute() (result map[string]interface{}, err error) {
	if m.metaTaskExec.isPutExtnV1B1Deploy() {
		result, err = m.putExtnV1B1Deploy()
	} else if m.metaTaskExec.isPutAppsV1B1Deploy() {
		result, err = m.putAppsV1B1Deploy()
	} else if m.metaTaskExec.isPatchExtnV1B1Deploy() {
		result, err = m.patchExtnV1B1Deploy()
	} else if m.metaTaskExec.isPatchAppsV1B1Deploy() {
		result, err = m.patchAppsV1B1Deploy()
	} else if m.metaTaskExec.isPutCoreV1Service() {
		result, err = m.putCoreV1Service()
	} else if m.metaTaskExec.isDeleteExtnV1B1Deploy() {
		result, err = m.deleteExtnV1B1Deployment()
	} else if m.metaTaskExec.isDeleteAppsV1B1Deploy() {
		result, err = m.deleteAppsV1B1Deployment()
	} else if m.metaTaskExec.isDeleteCoreV1Service() {
		result, err = m.deleteCoreV1Service()
	} else if m.metaTaskExec.isGetOEV1alpha1SP() {
		result, err = m.getOEV1alpha1SP()
	} else if m.metaTaskExec.isGetCoreV1PVC() {
		result, err = m.getCoreV1PVC()
	} else if m.metaTaskExec.isListCoreV1Pod() {
		result, err = m.listCoreV1Pod()
	} else {
		return nil, fmt.Errorf("Not a supported operation: '%#v'", m.metaTaskExec.getMetaInfo())
	}

	return result, err
}

// asRollback will provide the rollback instance w.r.t this task's instance
func (m *taskExecutor) asRollbackInstance(objectName string) (*taskExecutor, error) {
	mte, willRollback, err := m.metaTaskExec.asRollbackInstance(objectName)
	if err != nil {
		return nil, err
	}

	if !willRollback {
		// no need of rollback
		return nil, nil
	}

	kc, err := m_k8s_client.NewK8sClient(mte.getRunNamespace())
	if err != nil {
		return nil, err
	}

	// Only the meta info is required for a rollback. In
	// other words no need of task yaml template & values
	return &taskExecutor{
		objectName:   mte.getObjectName(),
		metaTaskExec: mte,
		k8sClient:    kc,
	}, nil
}

// asAppsV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *taskExecutor) asAppsV1B1Deploy() (*api_apps_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("AppsV1B1Deploy", m.task.yml, m.task.values)
	if err != nil {
		return nil, err
	}

	return d.AsAppsV1B1Deployment()
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *taskExecutor) asExtnV1B1Deploy() (*api_extn_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("ExtnV1B11Deploy", m.task.yml, m.task.values)
	if err != nil {
		return nil, err
	}

	return d.AsExtnV1B1Deployment()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *taskExecutor) asCoreV1Svc() (*api_core_v1.Service, error) {
	s, err := m_k8s.NewServiceYml("CoreV1Svc", m.task.yml, m.task.values)
	if err != nil {
		return nil, err
	}

	return s.AsCoreV1Service()
}

// putAppsV1B1Deploy will put a Deployment as defined in
// the Task
func (m *taskExecutor) putAppsV1B1Deploy() (map[string]interface{}, error) {
	d, err := m.asAppsV1B1Deploy()
	if err != nil {
		return nil, err
	}

	d, err = m.k8sClient.CreateAppsV1B1Deployment(d)
	if err != nil {
		return nil, err
	}

	// Set the results object
	r := map[string]interface{}{
		// set specific results with identity as the key
		m.identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): d.Name,
		},
		// set annotations with identity prefix as the key
		m.identity + "-" + string(v1alpha1.AnnotationsTRTP): d.Annotations,
	}

	return r, nil
}

// putExtnV1B1Deploy will put a Deployment as defined in
// the Task
func (m *taskExecutor) putExtnV1B1Deploy() (map[string]interface{}, error) {
	d, err := m.asExtnV1B1Deploy()
	if err != nil {
		return nil, err
	}

	d, err = m.k8sClient.CreateExtnV1B1Deployment(d)
	if err != nil {
		return nil, err
	}

	// Set the results object
	r := map[string]interface{}{
		// set specific results with identity as the key
		m.identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): d.Name,
		},
		// set annotations with identity prefix as the key
		m.identity + "-" + string(v1alpha1.AnnotationsTRTP): d.Annotations,
	}

	return r, nil
}

// patchAppsV1B1Deploy will patch a Deployment as defined in
// the Task
func (m *taskExecutor) patchAppsV1B1Deploy() (map[string]interface{}, error) {
	return nil, fmt.Errorf("Not Implemented")
}

// patchExtnV1B1Deploy will put a Deployment as defined in
// the Task
func (m *taskExecutor) patchExtnV1B1Deploy() (map[string]interface{}, error) {
	pe, err := newTaskPatchExecutor(m.taskPatch)
	if err != nil {
		return nil, err
	}

	pb, err := pe.build()
	if err != nil {
		return nil, err
	}

	deploy, err := m.k8sClient.PatchExtnV1B1DeploymentAsRaw(m.objectName, pe.patchType(), pb)
	if err != nil {
		return nil, err
	}

	e := newTaskResultQueryExecutor(m.identity, m.taskResultQueries, deploy)
	return e.execute()
}

// deleteAppsV1B1Deployment will delete a Deployment as defined in
// the Task
func (m *taskExecutor) deleteAppsV1B1Deployment() (map[string]interface{}, error) {
	return nil, m.k8sClient.DeleteAppsV1B1Deployment(m.objectName)
}

// deleteExtnV1B1Deployment will delete a Deployment as defined in
// the Task
func (m *taskExecutor) deleteExtnV1B1Deployment() (map[string]interface{}, error) {
	return nil, m.k8sClient.DeleteExtnV1B1Deployment(m.objectName)
}

// putCoreV1Service will put a Service as defined in
// the Task
func (m *taskExecutor) putCoreV1Service() (map[string]interface{}, error) {
	s, err := m.asCoreV1Svc()
	if err != nil {
		return nil, err
	}

	s, err = m.k8sClient.CreateCoreV1Service(s)
	if err != nil {
		return nil, err
	}

	// Set the resulting object & service ip
	// TODO Use hint e.g. jsonpath, etc in MetaTask to get:
	// Key i.e. serviceIP & Value i.e. s.Spec.ClusterIP
	r := map[string]interface{}{
		m.identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): s.Name,
			"serviceIP":                     s.Spec.ClusterIP,
		},
	}

	return r, nil
}

// deleteCoreV1Service will delete a Service as defined in
// the Task
func (m *taskExecutor) deleteCoreV1Service() (map[string]interface{}, error) {
	return nil, m.k8sClient.DeleteCoreV1Service(m.objectName)
}

// getOEV1alpha1SP will get the StoragePool as defined in the Task
func (m *taskExecutor) getOEV1alpha1SP() (map[string]interface{}, error) {
	sp, err := m.k8sClient.GetOEV1alpha1SP(m.objectName)
	if err != nil {
		return nil, err
	}

	// TODO Use hint e.g. jsonpath, etc in MetaTask to get
	// Key(s) & corresponding Value(s)
	r := map[string]interface{}{
		m.identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): sp.Name,
			"storagePoolPath":               sp.Spec.Path,
		},
	}

	return r, nil
}

// getCoreV1PVC will execute GET PVC API call. It will use the info
// available in the Task to execute this operation.
func (m *taskExecutor) getCoreV1PVC() (map[string]interface{}, error) {
	pvc, err := m.k8sClient.GetCoreV1PVCAsRaw(m.objectName)
	if err != nil {
		return nil, err
	}

	e := newTaskResultQueryExecutor(m.identity, m.taskResultQueries, pvc)
	return e.execute()
}

// listCoreV1Pod will execute List Pod API call. It will use the info
// available in the Task to execute this operation.
func (m *taskExecutor) listCoreV1Pod() (map[string]interface{}, error) {
	opts, err := m.metaTaskExec.getListOptions()
	if err != nil {
		return nil, err
	}

	listFn := func() (map[string]interface{}, error) {
		pods, err := m.k8sClient.ListCoreV1PodAsRaw(opts)
		if err != nil {
			return nil, err
		}

		e := newTaskResultQueryExecutor(m.identity, m.taskResultQueries, pods)
		return e.execute()
	}

	return m.retryOnVerificationError(listFn)
}

func (m *taskExecutor) retryOnVerificationError(fn func() (map[string]interface{}, error)) (op map[string]interface{}, err error) {
	attempts, interval := m.metaTaskExec.getRetry()

	for i := 0; i < attempts; i++ {
		op, err = fn()
		if err == nil {
			// return if successful func execution
			return
		}

		if _, ok := err.(*taskResultVerifyError); !ok {
			// return if not a verification error
			return
		}

		time.Sleep(interval)
	}
	// return after exhausting all attempts
	return
}

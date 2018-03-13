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
	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	m_k8s "github.com/openebs/maya/pkg/k8s"
	"github.com/openebs/maya/pkg/template"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// TaskAction signifies the action to be taken
// against a Task
type TaskAction string

const (
	// GetTA flags a action as get. Typically used to fetch
	// an object from its name.
	GetTA TaskAction = "get"
	// PutTA flags a action as put. Typically used to put
	// an object.
	PutTA TaskAction = "put"
	// DeleteTA flags a action as delete. Typically used to
	// delete an object.
	DeleteTA TaskAction = "delete"
)

// MetaTask contains information about a Task
type MetaTask struct {
	// Identifier provides a unique identification of this
	// task. There should not be two tasks with same identity.
	Identity string `json:"identity"`
	// Kind of the task
	Kind string `json:"kind"`
	// APIVersion of the task
	APIVersion string `json:"apiVersion"`
	// RunNamespace is the namespace where task will get
	// executed
	RunNamespace string `json:"runNamespace"`
	// Action to be invoked on the task
	Action TaskAction `json:"action"`
	// Owner represents the owner of this task
	Owner string `json:"owner"`
	// ObjectName is the name of the target that is
	// created or operated by the task
	ObjectName string `json:"objectName"`
	// TaskResultQueries will consist of the queries to be run against the
	// task's result
	TaskResultQueries []TaskResultQuery `json:"queries"`
	// QueryType flags the kind of query to be used to extract data from the
	// task's result
	//
	// NOTE:
	//  This may be taken up when such a need arises. e.g. flag to use either
	// Json Path, or Go Template, etc
	//QueryType QueryType `json:"queryType"`
}

// NewMetaTask provides a new instance of MetaTask from a
// yaml corresponding to MetaTask structure
func NewMetaTask(yml string, values map[string]interface{}) (*MetaTask, error) {
	// transform the yaml with provided values
	b, err := template.AsTemplatedBytes("MetaTask", yml, values)
	if err != nil {
		return nil, err
	}

	// unmarshall the yaml bytes into this instance
	var m MetaTask
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *MetaTask) isDeployment() bool {
	return m.Kind == string(m_k8s_client.DeploymentKK)
}

func (m *MetaTask) isService() bool {
	return m.Kind == string(m_k8s_client.ServiceKK)
}

func (m *MetaTask) isStoragePool() bool {
	return m.Kind == string(m_k8s_client.StroagePoolCRKK)
}

func (m *MetaTask) isConfigMap() bool {
	return m.Kind == string(m_k8s_client.ConfigMapKK)
}

func (m *MetaTask) isPVC() bool {
	return m.Kind == string(m_k8s_client.PersistentVolumeClaimKK)
}

func (m *MetaTask) isGet() bool {
	return m.Action == GetTA
}

func (m *MetaTask) isPut() bool {
	return m.Action == PutTA
}

func (m *MetaTask) isDelete() bool {
	return m.Action == DeleteTA
}

func (m *MetaTask) asRollbackInAction(objectName string) (*MetaTask, bool) {
	// there is no rollback when original action is not put
	if !m.isPut() {
		return nil, false
	}

	// A put action will translate to a delete action
	// keeping the objectName of the task same as the original
	mt := &MetaTask{
		Action:       DeleteTA,
		ObjectName:   objectName,
		Identity:     m.Identity,
		Kind:         m.Kind,
		APIVersion:   m.APIVersion,
		RunNamespace: m.RunNamespace,
		Owner:        m.Owner,
	}

	return mt, true
}

func (m *MetaTask) isExtnV1B1() bool {
	return m.APIVersion == string(m_k8s_client.ExtensionsV1Beta1KA)
}

func (m *MetaTask) isCoreV1() bool {
	return m.APIVersion == string(m_k8s_client.CoreV1KA)
}

func (m *MetaTask) isOEV1alpha1() bool {
	return m.APIVersion == string(m_k8s_client.OEV1alpha1KA)
}

func (m *MetaTask) isExtnV1B1Deploy() bool {
	return m.isExtnV1B1() && m.isDeployment()
}

func (m *MetaTask) isCoreV1Service() bool {
	return m.isCoreV1() && m.isService()
}

func (m *MetaTask) isCoreV1PVC() bool {
	return m.isCoreV1() && m.isPVC()
}

func (m *MetaTask) isOEV1alpha1SP() bool {
	return m.isOEV1alpha1() && m.isStoragePool()
}

func (m *MetaTask) isPutExtnV1B1Deploy() bool {
	return m.isExtnV1B1Deploy() && m.isPut()
}

func (m *MetaTask) isPutCoreV1Service() bool {
	return m.isCoreV1Service() && m.isPut()
}

func (m *MetaTask) isDeleteExtnV1B1Deploy() bool {
	return m.isExtnV1B1Deploy() && m.isDelete()
}

func (m *MetaTask) isDeleteCoreV1Service() bool {
	return m.isCoreV1Service() && m.isDelete()
}

func (m *MetaTask) isGetOEV1alpha1SP() bool {
	return m.isOEV1alpha1SP() && m.isGet()
}

// isGetCoreV1PVC flags if task is a GET action of
// PVC Kind
func (m *MetaTask) isGetCoreV1PVC() bool {
	return m.isCoreV1PVC() && m.isGet()
}

// Task represents a task that is capable of being
// executed in a workflow. A task execution typically
// means invoking an API call e.g. a K8s API call.
type Task struct {
	// MetaTask provides the information about this
	// task
	MetaTask
	// Values are the inputs that needs to be provided
	// to this task's template i.e. yaml
	values map[string]interface{}
	// yml represents the YAML representation of this task
	yml string
	// k8sClient will make K8s API calls
	// This is useful for mocking purposes
	k8sClient *m_k8s_client.K8sClient
}

// NewTask returns a new instance of Task
func NewTask(identity string, metaTaskYml, taskYml string, values map[string]interface{}) (*Task, error) {

	mt, err := NewMetaTask(metaTaskYml, values)
	if err != nil {
		return nil, err
	}

	// Give a unique identification for the task if not provided
	if len(mt.Identity) == 0 {
		mt.Identity = identity
	}

	kc, err := m_k8s_client.NewK8sClient(mt.RunNamespace)
	if err != nil {
		return nil, err
	}

	return &Task{
		MetaTask:  *mt,
		yml:       taskYml,
		values:    values,
		k8sClient: kc,
	}, nil
}

// Execute will execute the Task depending on informations
// available in Meta
func (m *Task) execute() (result map[string]interface{}, err error) {

	if m.isPutExtnV1B1Deploy() {
		result, err = m.putExtnV1B1Deploy()
	} else if m.isPutCoreV1Service() {
		result, err = m.putCoreV1Service()
	} else if m.isDeleteExtnV1B1Deploy() {
		result, err = m.deleteExtnV1B1Deployment()
	} else if m.isDeleteCoreV1Service() {
		result, err = m.deleteCoreV1Service()
	} else if m.isGetOEV1alpha1SP() {
		result, err = m.getOEV1alpha1SP()
	} else if m.isGetCoreV1PVC() {
		result, err = m.getCoreV1PVC()
	} else {
		return nil, fmt.Errorf("Not supported operation: '%#v'", m.MetaTask)
	}

	return result, err
}

// asRollback will provide the rollback instance w.r.t this task's instance
func (m *Task) asRollback(objectName string) (*Task, error) {
	mt, willRollback := m.MetaTask.asRollbackInAction(objectName)
	if !willRollback {
		return nil, nil
	}

	kc, err := m_k8s_client.NewK8sClient(mt.RunNamespace)
	if err != nil {
		return nil, err
	}

	// Only the meta info is required for a rollback. In
	// other words no need of yaml template & values
	return &Task{
		MetaTask:  *mt,
		k8sClient: kc,
	}, nil
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *Task) asExtnV1B1Deploy() (*api_extn_v1beta1.Deployment, error) {
	b, err := template.AsTemplatedBytes("ExtnV1beta1Deploy", m.yml, m.values)
	if err != nil {
		return nil, err
	}

	d := m_k8s.NewDeployment(b)
	return d.AsExtnV1B1Deployment()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *Task) asCoreV1Svc() (*api_core_v1.Service, error) {
	b, err := template.AsTemplatedBytes("CoreV1Svc", m.yml, m.values)
	if err != nil {
		return nil, err
	}

	s := m_k8s.NewService(b)
	return s.AsCoreV1Service()
}

// putExtnV1B1Deploy will put a Deployment as defined in
// the Task
func (m *Task) putExtnV1B1Deploy() (map[string]interface{}, error) {
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
		m.Identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): d.Name,
		},
		// set annotations with identity prefix as the key
		m.Identity + "-" + string(v1alpha1.AnnotationsTRTP): d.Annotations,
	}

	return r, nil
}

// deleteExtnV1B1Deployment will delete a Deployment as defined in
// the Task
func (m *Task) deleteExtnV1B1Deployment() (map[string]interface{}, error) {
	return nil, m.k8sClient.DeleteExtnV1B1Deployment(m.ObjectName)
}

// putCoreV1Service will put a Service as defined in
// the Task
func (m *Task) putCoreV1Service() (map[string]interface{}, error) {
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
		m.MetaTask.Identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): s.Name,
			"serviceIP":                     s.Spec.ClusterIP,
		},
	}

	return r, nil
}

// deleteCoreV1Service will delete a Service as defined in
// the Task
func (m *Task) deleteCoreV1Service() (map[string]interface{}, error) {
	return nil, m.k8sClient.DeleteCoreV1Service(m.ObjectName)
}

// getOEV1alpha1SP will get the StoragePool as defined in the Task
func (m *Task) getOEV1alpha1SP() (map[string]interface{}, error) {
	sp, err := m.k8sClient.GetOEV1alpha1SP(m.ObjectName)
	if err != nil {
		return nil, err
	}

	// TODO Use hint e.g. jsonpath, etc in MetaTask to get
	// Key(s) & corresponding Value(s)
	r := map[string]interface{}{
		m.MetaTask.Identity: map[string]string{
			string(v1alpha1.ObjectNameTRTP): sp.Name,
			"storagePoolPath":               sp.Spec.Path,
		},
	}

	return r, nil
}

// getCoreV1PVC will execute GET PVC API call. It will use the info
// available in the Task to execute this operation.
func (m *Task) getCoreV1PVC() (map[string]interface{}, error) {
	pvc, err := m.k8sClient.GetCoreV1PVCAsRaw(m.ObjectName)
	if err != nil {
		return nil, err
	}

	s := NewTaskResultStorage(m.MetaTask.Identity, m.TaskResultQueries, pvc)
	return s.store()
}

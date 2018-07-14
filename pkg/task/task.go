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
	"strings"
	"time"

	//"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	m_k8s "github.com/openebs/maya/pkg/k8s"
	"github.com/openebs/maya/pkg/template"
	"github.com/openebs/maya/pkg/util"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// TaskExecutor is the interface that provides a contract method to execute
// tasks
type TaskExecutor interface {
	Execute() (result map[string]interface{}, err error)
}

// TaskOutputExecutor is the interface that provides a contract method to
// generate output in a pre-defined format. The output format is specified in
// the task.
type TaskOutputExecutor interface {
	Output() (output []byte, err error)
}

// Task represents a task that is capable of being
// executed in a workflow.
//
// NOTE:
//  TaskGroupRunner is one such workflow runner.
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

// newTaskExecutor returns a new instance of taskExecutor
func newTaskExecutor(identity, metaTaskYml, taskYml string, values map[string]interface{}) (*taskExecutor, error) {
	mte, err := newMetaTaskExecutor(identity, metaTaskYml, values)
	if err != nil {
		return nil, err
	}

	// kubernetes clientset
	kubernetesClientSet, err := m_k8s_client.GetInClusterCS()
	if err != nil {
		return nil, err
	}

	// openEBS clientset
	openEBSClientSet, err := m_k8s_client.GetInClusterOECS()
	if err != nil {
		return nil, err
	}
	// client to make K8s API calls using the namespace on
	// which this task is supposed to be executed
	kc := m_k8s_client.NewK8sClient(kubernetesClientSet, openEBSClientSet, mte.getRunNamespace())

	return &taskExecutor{
		identity:          identity,
		objectName:        mte.getObjectName(),
		taskResultQueries: mte.getTaskResultQueries(),
		//taskPatch:         mte.getTaskPatch(),
		metaTaskExec: mte,
		task: Task{
			yml:    taskYml,
			values: values,
		},
		k8sClient: kc,
	}, nil
}

// resetK8sClient returns a new instance of taskExecutor pointing to a new
// namespace
func (m *taskExecutor) resetK8sClient(namespace string) (*taskExecutor, error) {
	// kubernetes clientset
	kubernetesClientSet, err := m_k8s_client.GetInClusterCS()
	if err != nil {
		return nil, err
	}

	// openEBS clientset
	openEBSClientSet, err := m_k8s_client.GetInClusterOECS()
	if err != nil {
		return nil, err
	}
	// client to make K8s API calls using the provided namespace
	kc := m_k8s_client.NewK8sClient(kubernetesClientSet, openEBSClientSet, namespace)
	// reset the k8s client
	m.k8sClient = kc
	return m, nil
}

// getMetaTaskExecutor gets the meta task executor value
func (m *taskExecutor) getMetaTaskExecutor() metaTaskExecutor {
	return *m.metaTaskExec
}

// Output returns the result of templating this task's yaml
//
// This implements TaskOutputExecutor interface
func (m *taskExecutor) Output() (output []byte, err error) {
	output, err = template.AsTemplatedBytes("Output", m.task.yml, m.task.values)
	return
}

// Execute will execute the task depending on task's meta information
//
// This implements TaskExecutor interface
func (m *taskExecutor) Execute() (result map[string]interface{}, err error) {
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
		result, err = m.listK8sResources()
	} else if m.metaTaskExec.isListCoreV1Service() {
		result, err = m.listK8sResources()
	} else if m.metaTaskExec.isListExtnV1B1Deploy() {
		result, err = m.listK8sResources()
	} else if m.metaTaskExec.isListAppsV1B1Deploy() {
		result, err = m.listK8sResources()
	} else {
		result = nil
		err = fmt.Errorf("failed to execute task: not a supported operation: meta info '%#v'", m.metaTaskExec.getMetaInfo())
	}

	return result, err
}

// asRollbackInstance will provide the rollback instance w.r.t this task's instance
func (m *taskExecutor) asRollbackInstance(objectName string) (*taskExecutor, error) {
	mte, willRollback, err := m.metaTaskExec.asRollbackInstance(objectName)
	if err != nil {
		return nil, err
	}

	if !willRollback {
		// no need of rollback
		return nil, nil
	}

	// kubernetes clientset
	kubernetesClientSet, err := m_k8s_client.GetInClusterCS()
	if err != nil {
		return nil, err
	}

	// openEBS clientset
	openEBSClientSet, err := m_k8s_client.GetInClusterOECS()
	if err != nil {
		return nil, err
	}

	kc := m_k8s_client.NewK8sClient(kubernetesClientSet, openEBSClientSet, mte.getRunNamespace())

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

	// put the deployment
	//d, err = m.k8sClient.CreateAppsV1B1Deployment(d)
	//if err != nil {
	//	return nil, err
	//}

	// Set the results object
	//r := map[string]interface{}{
	// set specific results with identity as the key
	//	m.identity: map[string]string{
	//		string(v1alpha1.ObjectNameTRTP): d.Name,
	//	},
	// set annotations with identity prefix as the key
	//	m.identity + "-" + string(v1alpha1.AnnotationsTRTP): d.Annotations,
	//}

	//return r, nil

	deploy, err := m.k8sClient.CreateAppsV1B1DeploymentAsRaw(d)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, deploy)
	return e.formattedResult()
}

// putExtnV1B1Deploy will put (i.e. apply to kubernetes cluster) a Deployment
// whose specifications are defined in the task and query/extract the specified
// properties from this i.e. resulting deployment object
func (m *taskExecutor) putExtnV1B1Deploy() (map[string]interface{}, error) {
	d, err := m.asExtnV1B1Deploy()
	if err != nil {
		return nil, err
	}

	// put the deployment
	//d, err = m.k8sClient.CreateExtnV1B1Deployment(d)
	//if err != nil {
	//return nil, err
	//}

	// Set the results object
	//r := map[string]interface{}{
	// set specific results with identity as the key
	//	m.identity: map[string]string{
	//		string(v1alpha1.ObjectNameTRTP): d.Name,
	//	},
	// set annotations with identity prefix as the key
	//	m.identity + "-" + string(v1alpha1.AnnotationsTRTP): d.Annotations,
	//}

	//return r, nil

	deploy, err := m.k8sClient.CreateExtnV1B1DeploymentAsRaw(d)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, deploy)
	return e.formattedResult()
}

// patchAppsV1B1Deploy will patch a Deployment as defined in the task
func (m *taskExecutor) patchAppsV1B1Deploy() (map[string]interface{}, error) {
	return nil, fmt.Errorf("patchAppsV1B1Deploy is not implemented")
}

// patchExtnV1B1Deploy will patch a Deployment as defined in the task
func (m *taskExecutor) patchExtnV1B1Deploy() (map[string]interface{}, error) {

	patch, err := asTaskPatch("ExtnV1B1DeployPatch", m.task.yml, m.task.values)
	if err != nil {
		return nil, err
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return nil, err
	}

	raw, err := pe.toJson()
	if err != nil {
		return nil, err
	}

	// patch the deployment
	deploy, err := m.k8sClient.PatchExtnV1B1DeploymentAsRaw(m.objectName, pe.patchType(), raw)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, deploy)
	return e.formattedResult()
}

// deleteAppsV1B1Deployment will delete one or more Deployments as defined in
// the task
func (m *taskExecutor) deleteAppsV1B1Deployment() (result map[string]interface{}, err error) {
	objectNames := strings.Split(strings.TrimSpace(m.objectName), ",")

	for _, name := range objectNames {
		err = m.k8sClient.DeleteAppsV1B1Deployment(name)
		if err != nil {
			return
		}
	}

	return
}

// deleteExtnV1B1Deployment will delete one or more Deployments as defined in
// the task
func (m *taskExecutor) deleteExtnV1B1Deployment() (result map[string]interface{}, err error) {
	objectNames := strings.Split(strings.TrimSpace(m.objectName), ",")

	for _, name := range objectNames {
		err = m.k8sClient.DeleteExtnV1B1Deployment(name)
		if err != nil {
			return
		}
	}

	return
}

// putCoreV1Service will put a Service as defined in the task
func (m *taskExecutor) putCoreV1Service() (map[string]interface{}, error) {
	s, err := m.asCoreV1Svc()
	if err != nil {
		return nil, err
	}

	//s, err = m.k8sClient.CreateCoreV1Service(s)
	//if err != nil {
	//	return nil, err
	//}

	// Set the resulting object & service ip
	// TODO Use hint e.g. jsonpath, etc in MetaTask to get:
	// Key i.e. serviceIP & Value i.e. s.Spec.ClusterIP
	//r := map[string]interface{}{
	//	m.identity: map[string]string{
	//		string(v1alpha1.ObjectNameTRTP): s.Name,
	//		"serviceIP":                     s.Spec.ClusterIP,
	//	},
	//}

	//return r, nil

	svc, err := m.k8sClient.CreateCoreV1ServiceAsRaw(s)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, svc)
	return e.formattedResult()
}

// deleteCoreV1Service will delete one or more services as defined in
// the task
func (m *taskExecutor) deleteCoreV1Service() (result map[string]interface{}, err error) {
	objectNames := strings.Split(strings.TrimSpace(m.objectName), ",")

	for _, name := range objectNames {
		err = m.k8sClient.DeleteCoreV1Service(name)
		if err != nil {
			return
		}
	}

	return
}

// getOEV1alpha1SP will get the StoragePool as defined in the task
func (m *taskExecutor) getOEV1alpha1SP() (map[string]interface{}, error) {
	//sp, err := m.k8sClient.GetOEV1alpha1SP(m.objectName)
	//if err != nil {
	//	return nil, err
	//}

	// TODO Use hint e.g. jsonpath, etc in MetaTask to get
	// Key(s) & corresponding Value(s)
	//r := map[string]interface{}{
	//	m.identity: map[string]string{
	//		string(v1alpha1.ObjectNameTRTP): sp.Name,
	//		"storagePoolPath":               sp.Spec.Path,
	//	},
	//}

	//return r, nil

	sp, err := m.k8sClient.GetOEV1alpha1SPAsRaw(m.objectName)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, sp)
	return e.formattedResult()
}

// getExtnV1B1Deployment will get the Deployment as defined in the task and then
// query for the specified properties from this Deployment object
func (m *taskExecutor) getExtnV1B1Deployment() (map[string]interface{}, error) {
	deploy, err := m.k8sClient.GetExtnV1B1DeploymentAsRaw(m.objectName)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, deploy)
	return e.formattedResult()
}

// getAppsV1B1Deployment will get the Deployment as defined in the task and then
// query for the specified properties from this Deployment object
func (m *taskExecutor) getAppsV1B1Deployment() (map[string]interface{}, error) {
	deploy, err := m.k8sClient.GetAppsV1B1DeploymentAsRaw(m.objectName)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, deploy)
	return e.formattedResult()
}

// getCoreV1PVC will get the PVC as defined in the task
func (m *taskExecutor) getCoreV1PVC() (map[string]interface{}, error) {
	pvc, err := m.k8sClient.GetCoreV1PVCAsRaw(m.objectName)
	if err != nil {
		return nil, err
	}

	e := newQueryExecFormatter(m.identity, m.taskResultQueries, pvc)
	return e.formattedResult()
}

// listK8sResources will list resources and extract each of these resource's
// specified properties
func (m *taskExecutor) listK8sResources() (map[string]interface{}, error) {
	opts, err := m.metaTaskExec.getListOptions()
	if err != nil {
		return nil, err
	}

	// list operation may deal with more than one namespaces
	runNS := m.metaTaskExec.getRunNamespace()
	runNamespaces := strings.Split(strings.TrimSpace(runNS), ",")
	list := map[string]interface{}{}
	nsCount := len(runNamespaces)

	// list for each namespace that is specified
	for _, ns := range runNamespaces {
		// closure that accepts a namespace to execute a list operation
		lFn := func(namespace string) (map[string]string, error) {
			var rs []byte
			var err error

			// change the k8s client namespace
			m.resetK8sClient(namespace)

			if m.metaTaskExec.isListCoreV1Pod() {
				rs, err = m.k8sClient.ListCoreV1PodAsRaw(opts)
			} else if m.metaTaskExec.isListCoreV1Service() {
				rs, err = m.k8sClient.ListCoreV1ServiceAsRaw(opts)
			} else if m.metaTaskExec.isListExtnV1B1Deploy() {
				rs, err = m.k8sClient.ListExtnV1B1DeploymentAsRaw(opts)
			} else if m.metaTaskExec.isListAppsV1B1Deploy() {
				rs, err = m.k8sClient.ListAppsV1B1DeploymentAsRaw(opts)
			} else {
				err = fmt.Errorf("failed to list k8s resources: meta task not supported '%#v'", m.metaTaskExec.getMetaInfo())
			}

			if err != nil {
				return nil, err
			}

			e := newQueryExecutor(m.taskResultQueries, rs)
			return e.result()
		}

		nsResult, err := m.retryOnVerificationError(ns, lFn)
		if err != nil {
			return nil, err
		}

		if nsCount == 1 {
			// set the results against the task id
			util.SetNestedField(list, nsResult, m.identity)
		} else {
			// set the results against the taskid.namespace
			util.SetNestedField(list, nsResult, m.identity, ns)
		}
	}

	return list, nil
}

func (m *taskExecutor) retryOnVerificationError(namespace string, fn func(string) (map[string]string, error)) (op map[string]string, err error) {
	attempts, interval := m.metaTaskExec.getRetry()

	for i := 0; i < attempts; i++ {
		op, err = fn(namespace)
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

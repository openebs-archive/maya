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

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
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
	Execute() (err error)
}

// TaskOutputExecutor is the interface that provides a contract method to
// generate output in a pre-defined format. The output format is specified in
// the task.
type TaskOutputExecutor interface {
	Output() (output []byte, err error)
}

// TODO
//  Refactor this to a Kubernetes Custom Resource
//
// RunTask composes various specifications of a task
type RunTask struct {
	// Name of this task
	Name string
	// MetaYml is the specifications about meta information of this run task
	MetaYml string
	// TaskYml is the specifications about this run task
	TaskYml string
	// PostRunTemplateFuncs is a set of go template functions that is run
	// against the result of this task's execution. In other words, this
	// template is run post the task execution.
	PostRunTemplateFuncs string
}

type taskExecutor struct {
	// identity is the id of the task
	identity string
	// templateValues will hold the values that will be applied against
	// the task's specification (which is a go template) before this task gets
	// executed
	templateValues map[string]interface{}
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
	// runtask is the specifications that determine a task & operations associated
	// with it
	runtask RunTask
	// k8sClient will be used to make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// newK8sClient returns a new instance of K8sClient based on the provided run
// namespace.
//
// NOTE:
//  Providing a run namespace can be optional. It is optional for cluster wide
// operations.
//
// NOTE:
//  In cases where more than one namespaces are involved, **repeatWith**
// metatask property is used.
func newK8sClient(runNamespace string) (kc *m_k8s_client.K8sClient, err error) {
	ns := strings.TrimSpace(runNamespace)
	kc, err = m_k8s_client.NewK8sClient(ns)
	return
}

// resetK8sClient returns a new instance of taskExecutor pointing to a new
// namespace
func (m *taskExecutor) resetK8sClient(namespace string) (*taskExecutor, error) {
	// client to make K8s API calls using the provided namespace
	kc, err := newK8sClient(namespace)
	if err != nil {
		return nil, err
	}

	// reset the k8s client
	m.k8sClient = kc
	return m, nil
}

// newTaskExecutor returns a new instance of taskExecutor
func newTaskExecutor(runtask RunTask, values map[string]interface{}) (*taskExecutor, error) {
	mte, err := newMetaTaskExecutor(runtask.MetaYml, values)
	if err != nil {
		return nil, err
	}

	// client to make K8s API calls using the namespace on
	// which this task is supposed to be executed
	kc, err := newK8sClient(mte.getRunNamespace())
	if err != nil {
		return nil, err
	}

	return &taskExecutor{
		templateValues:    values,
		identity:          mte.getIdentity(),
		objectName:        mte.getObjectName(),
		taskResultQueries: mte.getTaskResultQueries(),
		metaTaskExec:      mte,
		runtask:           runtask,
		k8sClient:         kc,
	}, nil
}

// String provides the essential task executor details
func (m *taskExecutor) String() string {
	return fmt.Sprintf("task with identity '%s' and with objectname '%s'", m.identity, m.objectName)
}

// getTaskIdentity gets the meta task executor value
func (m *taskExecutor) getTaskIdentity() string {
	return m.identity
}

// Output returns the result of templating this task's yaml
//
// This implements TaskOutputExecutor interface
func (m *taskExecutor) Output() (output []byte, err error) {
	output, err = template.AsTemplatedBytes("Output", m.runtask.TaskYml, m.templateValues)
	return
}

// getTaskResultNotFoundError fetches the NotFound error if any from this
// runtask's template values
//
// NOTE:
//  Logic to determine NotFound error is set at PostRunTemplateFuncs & is
// executed during post task execution phase. NotFound error is set if specified
// items or properties are not found. This error is set in the runtask's
// template values.
//
// NOTE:
//  Below property is set with verification error if any:
//  .TaskResult.<taskID>.notFoundErr
func (m *taskExecutor) getTaskResultNotFoundError() interface{} {
	return util.GetNestedField(m.templateValues, string(v1alpha1.TaskResultTLP), m.identity, string(v1alpha1.TaskResultNotFoundErrTRTP))
}

// getTaskResultVerifyError fetches the verification error if any from this
// runtask's template values
//
// NOTE:
//  Logic to determine Verify error is set at PostRunTemplateFuncs & is
// executed during post task execution phase. Verify error is set if specified
// verifications fail. This error is set in the runtask's template values.
//
// NOTE:
//  Below property is set with verification error if any:
//  .TaskResult.<taskID>.verifyErr
func (m *taskExecutor) getTaskResultVerifyError() interface{} {
	return util.GetNestedField(m.templateValues, string(v1alpha1.TaskResultTLP), m.identity, string(v1alpha1.TaskResultVerifyErrTRTP))
}

// resetTaskResultVerifyError resets the verification error from this runtask's
// template values
//
// NOTE:
//  reset here implies setting the verification err's placeholder value to nil
//
//  Below property is reset with `nil`:
//  .TaskResult.<taskID>.verifyErr
//
// NOTE:
//  Verification error is set during the post task execution phase if there are
// any verification error. This error is set in the runtask's template values.
func (m *taskExecutor) resetTaskResultVerifyError() {
	util.SetNestedField(m.templateValues, nil, string(v1alpha1.TaskResultTLP), m.identity, string(v1alpha1.TaskResultVerifyErrTRTP))
}

// repeatWith repeats execution of the task based on the repeatWith property
// set in meta task specifications. The same task is executed repeatedly based
// on the resource names set against the repeatWith property.
//
// NOTE:
//  Each task execution depends on the currently active repeat resource.
func (m *taskExecutor) repeatWith() (err error) {
	rwExec := m.metaTaskExec.getRepeatWithResourceExecutor()

	if !rwExec.isRepeat() {
		// no need to repeat if this task is not meant to be repeated;
		// so execute once & return
		err = m.retryOnVerificationError()
		return
	}

	// execute the task function based on the repeat resources
	for _, resource := range rwExec.getResources() {
		if rwExec.isNamespaceRepeat() {
			// if repetition is based on namespace, then the k8s client needs to
			// point to proper namespace before executing the task
			m.resetK8sClient(resource)
		}

		// set the currently active repeat resource
		util.SetNestedField(m.templateValues, resource, string(v1alpha1.ListItemsTLP), string(v1alpha1.CurrentRepeatResourceLITP))

		// execute the task function
		err = m.retryOnVerificationError()
		if err != nil {
			// stop repetition on unhandled runtime error & return
			return
		}
	}

	return
}

// retryOnVerificationError retries execution of the task if the task execution
// resulted into verification error. The number of retry attempts & interval
// between each attempt is specified in the task's meta specification.
func (m *taskExecutor) retryOnVerificationError() (err error) {
	retryAttempts, interval := m.metaTaskExec.getRetry()

	// original invocation as well as all retry attempts
	// i == 0 implies original task execute invocation
	// i > 0 implies a retry operation
	for i := 0; i <= retryAttempts; i++ {
		// first reset the previous verify error if any
		m.resetTaskResultVerifyError()

		// execute the task function
		err = m.ExecuteIt()
		if err != nil {
			// break this retry execution loop if there were any runtime errors
			return
		}

		// check for VerifyError if any
		//
		// NOTE:
		//  VerifyError is a handled runtime error which is handled via templating
		//
		// NOTE:
		//  retry is done only if VerifyError is thrown during post task
		// execution
		verifyErr := m.getTaskResultVerifyError()
		if verifyErr == nil {
			// no need to retry if task execution was a success & there was no
			// verification error found with the task result
			return
		}

		// current verify error
		err, _ = verifyErr.(*template.VerifyError)

		if i != retryAttempts {
			glog.Warningf("verify error was found during post runtask operations '%s': error '%#v': will retry task execution'%d'", m.identity, err, i+1)

			// will retry after the specified interval
			time.Sleep(interval)
		}
	}

	// return after exhausting the original invocation and all retries;
	// verification error of the final attempt will be returned here
	return
}

// Execute executes a runtask by following the directives specified in the
// runtask's meta specifications and other conditions like presence of VerifyErr
func (m *taskExecutor) Execute() (err error) {
	return m.repeatWith()
}

// postExecuteIt executes a go template against the provided template values.
// This is run after executing a task.
//
// NOTE:
//  This go template is a set of template functions that queries specified
// properties from the result due to the task's execution & storing it at
// placeholders within the **template values**. This is done to query these
// extracted values while executing later runtasks by providing these runtasks
// with the updated **template values**.
func (m *taskExecutor) postExecuteIt() (err error) {
	if len(m.runtask.PostRunTemplateFuncs) == 0 {
		// nothing needs to be done
		return
	}

	// post runtask operation
	_, err = template.AsTemplatedBytes("PostRunTemplateFuncs", m.runtask.PostRunTemplateFuncs, m.templateValues)
	if err != nil {
		return
	}

	// NotFound error is a handled runtime error. It is thrown during go template
	// execution & set is in the template values. This needs to checked and thrown
	// as an error.
	notFoundErr := m.getTaskResultNotFoundError()
	if notFoundErr != nil {
		glog.Warningf("notfound error during post runtask operations '%s': error '%#v'", m.identity, notFoundErr)

		err, _ = notFoundErr.(*template.NotFoundError)
	}

	return
}

// ExecuteIt will execute the runtask based on its meta specs & task specs
func (m *taskExecutor) ExecuteIt() (err error) {
	if m.k8sClient == nil {
		emsg := "failed to execute task: nil k8s client: verify if run namespace was available"
		glog.Errorf(fmt.Sprintf("%s: metatask '%#v'", emsg, m.metaTaskExec.getMetaInfo()))
		err = fmt.Errorf("%s: task '%s'", emsg, m.getTaskIdentity())
		return
	}

	if m.metaTaskExec.isPutExtnV1B1Deploy() {
		err = m.putExtnV1B1Deploy()
	} else if m.metaTaskExec.isPutAppsV1B1Deploy() {
		err = m.putAppsV1B1Deploy()
	} else if m.metaTaskExec.isPatchExtnV1B1Deploy() {
		err = m.patchExtnV1B1Deploy()
	} else if m.metaTaskExec.isPatchAppsV1B1Deploy() {
		err = m.patchAppsV1B1Deploy()
	} else if m.metaTaskExec.isPutCoreV1Service() {
		err = m.putCoreV1Service()
	} else if m.metaTaskExec.isDeleteExtnV1B1Deploy() {
		err = m.deleteExtnV1B1Deployment()
	} else if m.metaTaskExec.isDeleteAppsV1B1Deploy() {
		err = m.deleteAppsV1B1Deployment()
	} else if m.metaTaskExec.isDeleteCoreV1Service() {
		err = m.deleteCoreV1Service()
	} else if m.metaTaskExec.isGetOEV1alpha1SP() {
		err = m.getOEV1alpha1SP()
	} else if m.metaTaskExec.isGetCoreV1PVC() {
		err = m.getCoreV1PVC()
	} else if m.metaTaskExec.isList() {
		err = m.listK8sResources()
	} else {
		err = fmt.Errorf("failed to execute task: not a supported operation: meta info '%#v'", m.metaTaskExec.getMetaInfo())
	}

	if err != nil {
		return
	}

	// run the post operations after a runtask is executed
	return m.postExecuteIt()
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

	kc, err := m_k8s_client.NewK8sClient(mte.getRunNamespace())
	if err != nil {
		return nil, err
	}

	// Only the meta info is required for a rollback. In
	// other words no need of task yaml template & values
	return &taskExecutor{
		identity:     m.identity,
		objectName:   mte.getObjectName(),
		metaTaskExec: mte,
		k8sClient:    kc,
	}, nil
}

// asAppsV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *taskExecutor) asAppsV1B1Deploy() (*api_apps_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("AppsV1B1Deploy", m.runtask.TaskYml, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsAppsV1B1Deployment()
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *taskExecutor) asExtnV1B1Deploy() (*api_extn_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("ExtnV1B11Deploy", m.runtask.TaskYml, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsExtnV1B1Deployment()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *taskExecutor) asCoreV1Svc() (*api_core_v1.Service, error) {
	s, err := m_k8s.NewServiceYml("CoreV1Svc", m.runtask.TaskYml, m.templateValues)
	if err != nil {
		return nil, err
	}

	return s.AsCoreV1Service()
}

// putAppsV1B1Deploy will put (i.e. apply to a kubernetes cluster) a Deployment
// object. The Deployment specs is configured in the RunTask.
func (m *taskExecutor) putAppsV1B1Deploy() (err error) {
	d, err := m.asAppsV1B1Deploy()
	if err != nil {
		return
	}

	deploy, err := m.k8sClient.CreateAppsV1B1DeploymentAsRaw(d)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// putExtnV1B1Deploy will put (i.e. apply to kubernetes cluster) a Deployment
// whose specifications are defined in the RunTask
func (m *taskExecutor) putExtnV1B1Deploy() (err error) {
	d, err := m.asExtnV1B1Deploy()
	if err != nil {
		return
	}

	deploy, err := m.k8sClient.CreateExtnV1B1DeploymentAsRaw(d)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// patchAppsV1B1Deploy will patch a Deployment object in a kubernetes cluster.
// The patch specifications as configured in the RunTask
func (m *taskExecutor) patchAppsV1B1Deploy() (err error) {
	err = fmt.Errorf("patchAppsV1B1Deploy is not implemented")
	return
}

// patchExtnV1B1Deploy will patch a Deployment where the patch specifications
// are configured in the RunTask
func (m *taskExecutor) patchExtnV1B1Deploy() (err error) {
	patch, err := asTaskPatch("ExtnV1B1DeployPatch", m.runtask.TaskYml, m.templateValues)
	if err != nil {
		return
	}

	pe, err := newTaskPatchExecutor(patch)
	if err != nil {
		return
	}

	raw, err := pe.toJson()
	if err != nil {
		return
	}

	// patch the deployment
	deploy, err := m.k8sClient.PatchExtnV1B1DeploymentAsRaw(m.objectName, pe.patchType(), raw)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// deleteAppsV1B1Deployment will delete one or more Deployments as specified in
// the RunTask
func (m *taskExecutor) deleteAppsV1B1Deployment() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.objectName), ",")

	for _, name := range objectNames {
		err = m.k8sClient.DeleteAppsV1B1Deployment(name)
		if err != nil {
			return
		}
	}

	return
}

// deleteExtnV1B1Deployment will delete one or more Deployments as specified in
// the RunTask
func (m *taskExecutor) deleteExtnV1B1Deployment() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.objectName), ",")

	for _, name := range objectNames {
		err = m.k8sClient.DeleteExtnV1B1Deployment(name)
		if err != nil {
			return
		}
	}

	return
}

// putCoreV1Service will put a Service whose specs are configured in the RunTask
func (m *taskExecutor) putCoreV1Service() (err error) {
	s, err := m.asCoreV1Svc()
	if err != nil {
		return
	}

	svc, err := m.k8sClient.CreateCoreV1ServiceAsRaw(s)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, svc, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// deleteCoreV1Service will delete one or more services as specified in
// the RunTask
func (m *taskExecutor) deleteCoreV1Service() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.objectName), ",")

	for _, name := range objectNames {
		err = m.k8sClient.DeleteCoreV1Service(name)
		if err != nil {
			return
		}
	}

	return
}

// getOEV1alpha1SP will get the StoragePool as specified in the RunTask
func (m *taskExecutor) getOEV1alpha1SP() (err error) {
	sp, err := m.k8sClient.GetOEV1alpha1SPAsRaw(m.objectName)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, sp, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// getExtnV1B1Deployment will get the Deployment as specified in the RunTask
func (m *taskExecutor) getExtnV1B1Deployment() (err error) {
	deploy, err := m.k8sClient.GetExtnV1B1DeploymentAsRaw(m.objectName)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// getAppsV1B1Deployment will get the Deployment as specified in the RunTask
func (m *taskExecutor) getAppsV1B1Deployment() (err error) {
	deploy, err := m.k8sClient.GetAppsV1B1DeploymentAsRaw(m.objectName)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// getCoreV1PVC will get the PVC as specified in the RunTask
func (m *taskExecutor) getCoreV1PVC() (err error) {
	pvc, err := m.k8sClient.GetCoreV1PVCAsRaw(m.objectName)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, pvc, string(v1alpha1.CurrentJsonResultTLP))
	return
}

// listK8sResources will list resources as specified in the RunTask
func (m *taskExecutor) listK8sResources() (err error) {
	opts, err := m.metaTaskExec.getListOptions()
	if err != nil {
		return
	}

	var op []byte

	if m.metaTaskExec.isListCoreV1Pod() {
		op, err = m.k8sClient.ListCoreV1PodAsRaw(opts)
	} else if m.metaTaskExec.isListCoreV1Service() {
		op, err = m.k8sClient.ListCoreV1ServiceAsRaw(opts)
	} else if m.metaTaskExec.isListExtnV1B1Deploy() {
		op, err = m.k8sClient.ListExtnV1B1DeploymentAsRaw(opts)
	} else if m.metaTaskExec.isListAppsV1B1Deploy() {
		op, err = m.k8sClient.ListAppsV1B1DeploymentAsRaw(opts)
	} else if m.metaTaskExec.isListCoreV1PVC() {
		op, err = m.k8sClient.ListCoreV1PVCAsRaw(opts)
	} else {
		err = fmt.Errorf("failed to list k8s resources: meta task not supported: task details '%#v'", m.metaTaskExec.getTaskIdentity())
	}

	if err != nil {
		return
	}

	// set the json doc result
	util.SetNestedField(m.templateValues, op, string(v1alpha1.CurrentJsonResultTLP))
	return
}

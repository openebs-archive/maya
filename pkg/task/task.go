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
	deploy_appsv1 "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	deploy_extnv1beta1 "github.com/openebs/maya/pkg/kubernetes/deployment/extnv1beta1/v1alpha1"
	podexec "github.com/openebs/maya/pkg/kubernetes/podexec/v1alpha1"
	"github.com/openebs/maya/pkg/template"
	"github.com/openebs/maya/pkg/util"
	api_apps_v1 "k8s.io/api/apps/v1"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_batch_v1 "k8s.io/api/batch/v1"
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

type taskExecutor struct {
	// templateValues will hold the values that will be applied against
	// the task's specification (which is a go template) before this task gets
	// executed
	templateValues map[string]interface{}

	// metaTaskExec is the instance to be used to execute meta
	// operations on this task
	metaTaskExec *metaTaskExecutor

	// runtask is the specifications that determine a task & operations associated
	// with it
	runtask *v1alpha1.RunTask
}

// newTaskExecutor returns a new instance of taskExecutor
func newTaskExecutor(runtask *v1alpha1.RunTask, values map[string]interface{}) (*taskExecutor, error) {
	mte, err := newMetaTaskExecutor(runtask.Spec.Meta, values)
	if err != nil {
		return nil, err
	}

	return &taskExecutor{
		templateValues: values,
		metaTaskExec:   mte,
		runtask:        runtask,
	}, nil
}

// String provides the essential task executor details
func (m *taskExecutor) String() string {
	return fmt.Sprintf("task with identity '%s' and with objectname '%s'", m.getTaskIdentity(), m.getTaskObjectName())
}

// getTaskIdentity gets the task identity
func (m *taskExecutor) getTaskIdentity() string {
	return m.metaTaskExec.getIdentity()
}

// getTaskObjectName gets the task's object name
func (m *taskExecutor) getTaskObjectName() string {
	return m.metaTaskExec.getObjectName()
}

// getTaskRunNamespace gets the task's run namespace
func (m *taskExecutor) getTaskRunNamespace() string {
	return m.metaTaskExec.getRunNamespace()
}

// getK8sClient gets the kubernetes client to execute this task
func (m *taskExecutor) getK8sClient() *m_k8s_client.K8sClient {
	return m.metaTaskExec.getK8sClient()
}

// Output returns the result of templating this task's yaml
//
// This implements TaskOutputExecutor interface
func (m *taskExecutor) Output() (output []byte, err error) {
	output, err = template.AsTemplatedBytes("Output", m.runtask.Spec.Task, m.templateValues)
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
	return util.GetNestedField(m.templateValues, string(v1alpha1.TaskResultTLP), m.getTaskIdentity(), string(v1alpha1.TaskResultNotFoundErrTRTP))
}

// getTaskResultVersionMismatchError fetches the VersionMismatch error if any
// from this runtask's template values
//
// NOTE:
//  Logic to determine VersionMismatch error is set at PostRunTemplateFuncs & is
// executed during post task execution phase. VersionMismatch error is set if
// task is executed for invalid version of the resource. This error is set in
// the runtask's template values.
//
// NOTE:
//  Below property is set with VersionMismatch error if any:
//  .TaskResult.<taskID>.versionMismatchErr
func (m *taskExecutor) getTaskResultVersionMismatchError() interface{} {
	return util.GetNestedField(m.templateValues, string(v1alpha1.TaskResultTLP), m.getTaskIdentity(), string(v1alpha1.TaskResultVersionMismatchErrTRTP))
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
	return util.GetNestedField(m.templateValues, string(v1alpha1.TaskResultTLP), m.getTaskIdentity(), string(v1alpha1.TaskResultVerifyErrTRTP))
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
	util.SetNestedField(m.templateValues, nil, string(v1alpha1.TaskResultTLP), m.getTaskIdentity(), string(v1alpha1.TaskResultVerifyErrTRTP))
}

// repeatWith repeats execution of the task based on the repeatWith property
// set in meta task specifications. The same task is executed repeatedly based
// on the resource names set against the repeatWith property.
//
// NOTE:
//  Each task execution depends on the currently active repeat resource.
func (m *taskExecutor) repeatWith() (err error) {
	rwExec := m.metaTaskExec.getRepeatExecutor()
	if !rwExec.isRepeat() {
		// no need to repeat if this task is not meant to be repeated;
		// so execute once & return
		err = m.retryOnVerificationError()
		return
	}

	// execute the task based on each repeat
	repeats := rwExec.len()
	var (
		rptMetaTaskExec *metaTaskExecutor
		current         string
	)

	for idx := 0; idx < repeats; idx++ {
		// fetch a new repeat meta task instance
		rptMetaTaskExec, err = m.metaTaskExec.asRepeatInstance(idx)
		if err != nil {
			// stop repetition on unhandled runtime errors & return
			return
		}
		// mutate the original meta task executor to this repeater instance
		m.metaTaskExec = rptMetaTaskExec

		// set the currently active repeat item
		current, err = m.metaTaskExec.repeater.getItem(idx)
		if err != nil {
			// stop repetition on unhandled runtime error & return
			return
		}
		util.SetNestedField(m.templateValues, current, string(v1alpha1.ListItemsTLP), string(v1alpha1.CurrentRepeatResourceLITP))

		// execute the task function... finally
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
			glog.Warningf("verify error was found during post runtask operations '%s': error '%+v': will retry task execution'%d'", m.getTaskIdentity(), err, i+1)

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
	if m.metaTaskExec.isDisabled() {
		return
	}
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
	if m.runtask == nil || len(m.runtask.Spec.PostRun) == 0 {
		// nothing needs to be done
		return
	}

	// post runtask operation
	_, err = template.AsTemplatedBytes("PostRun", m.runtask.Spec.PostRun, m.templateValues)
	if err != nil {
		// return any un-handled runtime error
		return
	}

	// verMismatchErr is a handled runtime error. It is thrown & handled during go
	// template execution & set is in the template values. This needs to be
	// extracted from template values and thrown as VersionMismatchError
	verMismatchErr := m.getTaskResultVersionMismatchError()
	if verMismatchErr != nil {
		glog.Warningf("version mismatch error during post runtask operations '%s': error '%+v'", m.getTaskIdentity(), verMismatchErr)
		err, _ = verMismatchErr.(*template.VersionMismatchError)
		return
	}

	// notFoundErr is a handled runtime error. It is thrown & handled during go
	// template execution & set is in the template values. This needs to be
	// extracted and thrown as NotFoundError
	notFoundErr := m.getTaskResultNotFoundError()
	if notFoundErr != nil {
		glog.Warningf("notfound error during post runtask operations '%s': error '%+v'", m.getTaskIdentity(), notFoundErr)
		err, _ = notFoundErr.(*template.NotFoundError)
		return
	}

	return nil
}

// ExecuteIt will execute the runtask based on its meta specs & task specs
func (m *taskExecutor) ExecuteIt() (err error) {
	if m.getK8sClient() == nil {
		emsg := "failed to execute task: nil k8s client: verify if run namespace was available"
		glog.Errorf(fmt.Sprintf("%s: metatask '%+v'", emsg, m.metaTaskExec.getMetaInfo()))
		err = fmt.Errorf("%s: task '%s'", emsg, m.getTaskIdentity())
		return
	}

	// kind as command is a special case of task execution
	if m.metaTaskExec.isCommand() {
		return m.postExecuteIt()
	}

	if m.metaTaskExec.isRolloutstatus() {
		err = m.rolloutStatus()
	} else if m.metaTaskExec.isPutExtnV1B1Deploy() {
		err = m.putExtnV1B1Deploy()
	} else if m.metaTaskExec.isPutAppsV1B1Deploy() {
		err = m.putAppsV1B1Deploy()
	} else if m.metaTaskExec.isPatchExtnV1B1Deploy() {
		err = m.patchExtnV1B1Deploy()
	} else if m.metaTaskExec.isPatchAppsV1B1Deploy() {
		err = m.patchAppsV1B1Deploy()
	} else if m.metaTaskExec.isPatchOEV1alpha1SPC() {
		err = m.patchOEV1alpha1SPC()
	} else if m.metaTaskExec.isPutCoreV1Service() {
		err = m.putCoreV1Service()
	} else if m.metaTaskExec.isPatchCoreV1Service() {
		err = m.patchCoreV1Service()
	} else if m.metaTaskExec.isDeleteExtnV1B1Deploy() {
		err = m.deleteExtnV1B1Deployment()
	} else if m.metaTaskExec.isDeleteExtnV1B1ReplicaSet() {
		err = m.deleteExtnV1B1ReplicaSet()
	} else if m.metaTaskExec.isGetExtnV1B1Deploy() {
		err = m.getExtnV1B1Deployment()
	} else if m.metaTaskExec.isDeleteAppsV1B1Deploy() {
		err = m.deleteAppsV1B1Deployment()
	} else if m.metaTaskExec.isDeleteCoreV1Service() {
		err = m.deleteCoreV1Service()
	} else if m.metaTaskExec.isGetOEV1alpha1Disk() {
		err = m.getOEV1alpha1Disk()
	} else if m.metaTaskExec.isGetOEV1alpha1SPC() {
		err = m.getOEV1alpha1SPC()
	} else if m.metaTaskExec.isGetOEV1alpha1SP() {
		err = m.getOEV1alpha1SP()
	} else if m.metaTaskExec.isGetCoreV1PVC() {
		err = m.getCoreV1PVC()
	} else if m.metaTaskExec.isPutOEV1alpha1CSP() {
		err = m.putCStorPool()
	} else if m.metaTaskExec.isPutOEV1alpha1SP() {
		err = m.putStoragePool()
	} else if m.metaTaskExec.isPutOEV1alpha1CSV() {
		err = m.putCStorVolume()
	} else if m.metaTaskExec.isPutOEV1alpha1CVR() {
		err = m.putCStorVolumeReplica()
	} else if m.metaTaskExec.isDeleteOEV1alpha1SP() {
		err = m.deleteOEV1alpha1SP()
	} else if m.metaTaskExec.isDeleteOEV1alpha1CSP() {
		err = m.deleteOEV1alpha1CSP()
	} else if m.metaTaskExec.isDeleteOEV1alpha1CSV() {
		err = m.deleteOEV1alpha1CSV()
	} else if m.metaTaskExec.isDeleteOEV1alpha1CVR() {
		err = m.deleteOEV1alpha1CVR()
	} else if m.metaTaskExec.isPatchOEV1alpha1CSV() {
		err = m.patchOEV1alpha1CSV()
	} else if m.metaTaskExec.isPatchOEV1alpha1CVR() {
		err = m.patchOEV1alpha1CVR()
	} else if m.metaTaskExec.isList() {
		err = m.listK8sResources()
	} else if m.metaTaskExec.isGetStorageV1SC() {
		err = m.getStorageV1SC()
	} else if m.metaTaskExec.isGetCoreV1PV() {
		err = m.getCoreV1PV()
	} else if m.metaTaskExec.isDeleteBatchV1Job() {
		err = m.deleteBatchV1Job()
	} else if m.metaTaskExec.isGetBatchV1Job() {
		err = m.getBatchV1Job()
	} else if m.metaTaskExec.isPutBatchV1Job() {
		err = m.putBatchV1Job()
	} else if m.metaTaskExec.isPutAppsV1STS() {
		err = m.putAppsV1STS()
	} else if m.metaTaskExec.isDeleteAppsV1STS() {
		err = m.deleteAppsV1STS()
	} else if m.metaTaskExec.isExecCoreV1Pod() {
		err = m.execCoreV1Pod()
	} else {
		err = fmt.Errorf("un-supported task operation: failed to execute task: '%+v'", m.metaTaskExec.getMetaInfo())
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

	// Only the meta info is required for a rollback. In
	// other words no need of task yaml template & values
	return &taskExecutor{
		metaTaskExec: mte,
	}, nil
}

// asBatchV1Job generates a K8s Job object out of the embedded yaml
func (m *taskExecutor) asBatchV1Job() (*api_batch_v1.Job, error) {
	j, err := m_k8s.NewJobYml("BatchV1Job", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}
	return j.AsBatchV1Job()
}

// asAppsV1STS generates a kubernetes StatefulSet api
// instance from the yaml string specification
func (m *taskExecutor) asAppsV1STS() (*api_apps_v1.StatefulSet, error) {
	s, err := m_k8s.NewSTSYml("AppsV1StatefulSet", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}
	return s.AsAppsV1STS()
}

// asAppsV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *taskExecutor) asAppsV1B1Deploy() (*api_apps_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("AppsV1B1Deploy", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsAppsV1B1Deployment()
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *taskExecutor) asExtnV1B1Deploy() (*api_extn_v1beta1.Deployment, error) {
	d, err := m_k8s.NewDeploymentYml("ExtnV1B11Deploy", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsExtnV1B1Deployment()
}

// asCStorPool generates a CstorPool object
// out of the embedded yaml
func (m *taskExecutor) asCStorPool() (*v1alpha1.CStorPool, error) {
	d, err := m_k8s.NewCStorPoolYml("CStorPool", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsCStorPoolYml()
}

// asStoragePool generates a StoragePool object
// out of the embedded yaml
func (m *taskExecutor) asStoragePool() (*v1alpha1.StoragePool, error) {
	d, err := m_k8s.NewStoragePoolYml("StoragePool", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsStoragePoolYml()
}

// asCStorVolume generates a CstorVolume object
// out of the embedded yaml
func (m *taskExecutor) asCStorVolume() (*v1alpha1.CStorVolume, error) {
	d, err := m_k8s.NewCStorVolumeYml("CstorVolume", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsCStorVolumeYml()
}

// asCstorVolumeReplica generates a CStorVolumeReplica object
// out of the embedded yaml
func (m *taskExecutor) asCstorVolumeReplica() (*v1alpha1.CStorVolumeReplica, error) {
	d, err := m_k8s.NewCStorVolumeReplicaYml("CstorVolumeReplica", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return d.AsCStorVolumeReplicaYml()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *taskExecutor) asCoreV1Svc() (*api_core_v1.Service, error) {
	s, err := m_k8s.NewServiceYml("CoreV1Svc", m.runtask.Spec.Task, m.templateValues)
	if err != nil {
		return nil, err
	}

	return s.AsCoreV1Service()
}

// putBatchV1Job will put a Job object
func (m *taskExecutor) putBatchV1Job() (err error) {
	j, err := m.asBatchV1Job()
	if err != nil {
		return
	}
	job, err := m.getK8sClient().CreateBatchV1JobAsRaw(j)
	if err != nil {
		return
	}
	util.SetNestedField(m.templateValues, job, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putAppsV1STS will create a new StatefulSet
// object in the cluster and store the response
// in a json format
func (m *taskExecutor) putAppsV1STS() (err error) {
	j, err := m.asAppsV1STS()
	if err != nil {
		return
	}
	sts, err := m.getK8sClient().CreateAppsV1STSAsRaw(j)
	if err != nil {
		return
	}
	util.SetNestedField(m.templateValues, sts, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putAppsV1B1Deploy will put (i.e. apply to a kubernetes cluster) a Deployment
// object. The Deployment specs is configured in the RunTask.
func (m *taskExecutor) putAppsV1B1Deploy() (err error) {
	d, err := m.asAppsV1B1Deploy()
	if err != nil {
		return
	}

	deploy, err := m.getK8sClient().CreateAppsV1B1DeploymentAsRaw(d)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putExtnV1B1Deploy will put (i.e. apply to kubernetes cluster) a Deployment
// whose specifications are defined in the RunTask
func (m *taskExecutor) putExtnV1B1Deploy() (err error) {
	d, err := m.asExtnV1B1Deploy()
	if err != nil {
		return
	}

	deploy, err := m.getK8sClient().CreateExtnV1B1DeploymentAsRaw(d)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// patchSPC will patch a SPC object in a kubernetes cluster.
// The patch specifications as configured in the RunTask
func (m *taskExecutor) patchOEV1alpha1SPC() (err error) {
	patch, err := asTaskPatch("patchSPC", m.runtask.Spec.Task, m.templateValues)
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

	// patch the SPC
	spc, err := m.getK8sClient().PatchOEV1alpha1SPCAsRaw(m.getTaskObjectName(), pe.patchType(), raw)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, spc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// patchOEV1alpha1CSV will patch a CStorVolume as defined in the task
func (m *taskExecutor) patchOEV1alpha1CSV() (err error) {
	patch, err := asTaskPatch("patchCSV", m.runtask.Spec.Task, m.templateValues)
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
	// patch the CStor Volume
	csv, err := m.getK8sClient().PatchOEV1alpha1CSV(m.getTaskObjectName(), m.getTaskRunNamespace(), pe.patchType(), raw)
	if err != nil {
		return
	}
	util.SetNestedField(m.templateValues, csv, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// patchOEV1alpha1CVR will patch a CStorVolumeReplica as defined in the task
func (m *taskExecutor) patchOEV1alpha1CVR() (err error) {
	patch, err := asTaskPatch("patchCVR", m.runtask.Spec.Task, m.templateValues)
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
	// patch the CStor Volume Replica
	cvr, err := m.getK8sClient().PatchOEV1alpha1CVR(m.getTaskObjectName(), m.getTaskRunNamespace(), pe.patchType(), raw)
	if err != nil {
		return
	}
	util.SetNestedField(m.templateValues, cvr, string(v1alpha1.CurrentJSONResultTLP))
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
	patch, err := asTaskPatch("ExtnV1B1DeployPatch", m.runtask.Spec.Task, m.templateValues)
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
	deploy, err := m.getK8sClient().PatchExtnV1B1DeploymentAsRaw(m.getTaskObjectName(), pe.patchType(), raw)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, deploy, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// patchCoreV1Service will patch a Service where the patch specifications
// are configured in the RunTask
func (m *taskExecutor) patchCoreV1Service() (err error) {
	patch, err := asTaskPatch("CoreV1ServicePatch", m.runtask.Spec.Task, m.templateValues)
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
	// patch the service
	service, err := m.getK8sClient().PatchCoreV1ServiceAsRaw(m.getTaskObjectName(), pe.patchType(), raw)
	if err != nil {
		return
	}
	util.SetNestedField(m.templateValues, service, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// deleteAppsV1B1Deployment will delete one or more Deployments as specified in
// the RunTask
func (m *taskExecutor) deleteAppsV1B1Deployment() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteAppsV1B1Deployment(strings.TrimSpace(name))
		if err != nil {
			return
		}
	}

	return
}

// deleteOEV1alpha1CVR will delete one or more CStorVolumeReplica as specified in
// the RunTask
func (m *taskExecutor) deleteOEV1alpha1CVR() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1CVR(name)
		if err != nil {
			return
		}
	}

	return
}

// deleteExtnV1B1Deployment will delete one or more Deployments as specified in
// the RunTask
func (m *taskExecutor) deleteExtnV1B1Deployment() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteExtnV1B1Deployment(strings.TrimSpace(name))
		if err != nil {
			return
		}
	}

	return
}

// deleteExtnV1B1ReplicaSet will delete one or more ReplicaSets as specified in
// the RunTask
func (m *taskExecutor) deleteExtnV1B1ReplicaSet() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")
	for _, name := range objectNames {
		err = m.getK8sClient().DeleteExtnV1B1ReplicaSet(strings.TrimSpace(name))
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

	svc, err := m.getK8sClient().CreateCoreV1ServiceAsRaw(s)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, svc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// deleteCoreV1Service will delete one or more services as specified in
// the RunTask
func (m *taskExecutor) deleteCoreV1Service() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteCoreV1Service(strings.TrimSpace(name))
		if err != nil {
			return
		}
	}

	return
}

// getOEV1alpha1Disk() will get the Disk as specified in the RunTask
func (m *taskExecutor) getOEV1alpha1Disk() (err error) {
	disk, err := m.getK8sClient().GetOEV1alpha1DiskAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, disk, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getOEV1alpha1SPC() will get the StoragePoolClaim as specified in the RunTask
func (m *taskExecutor) getOEV1alpha1SPC() (err error) {
	spc, err := m.getK8sClient().GetOEV1alpha1SPCAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, spc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getOEV1alpha1SP will get the StoragePool as specified in the RunTask
func (m *taskExecutor) getOEV1alpha1SP() (err error) {
	sp, err := m.getK8sClient().GetOEV1alpha1SPAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, sp, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getExtnV1B1Deployment will get the Deployment as specified in the RunTask
func (m *taskExecutor) getExtnV1B1Deployment() (err error) {
	dclient := deploy_extnv1beta1.KubeClient(
		deploy_extnv1beta1.WithNamespace(m.getTaskRunNamespace()),
		deploy_extnv1beta1.WithClientset(m.getK8sClient().GetKCS()))
	d, err := dclient.Get(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, d, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// extnV1B1DeploymentRollOutStatus generates rollout status for a given deployment from deployment object
func (m *taskExecutor) extnV1B1DeploymentRollOutStatus() (err error) {
	dclient := deploy_extnv1beta1.KubeClient(
		deploy_extnv1beta1.WithNamespace(m.getTaskRunNamespace()),
		deploy_extnv1beta1.WithClientset(m.getK8sClient().GetKCS()))
	res, err := dclient.RolloutStatusf(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, res, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getAppsV1DeploymentRollOutStatus generates rollout status for a given deployment from deployment object
func (m *taskExecutor) appsV1DeploymentRollOutStatus() (err error) {
	dclient := deploy_appsv1.KubeClient(
		deploy_appsv1.WithNamespace(m.getTaskRunNamespace()),
		deploy_appsv1.WithClientset(m.getK8sClient().GetKCS()))
	res, err := dclient.RolloutStatusf(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, res, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getAppsV1Deployment will get the Deployment as specified in the RunTask
func (m *taskExecutor) getAppsV1Deployment() (err error) {
	dclient := deploy_appsv1.KubeClient(
		deploy_appsv1.WithNamespace(m.getTaskRunNamespace()),
		deploy_appsv1.WithClientset(m.getK8sClient().GetKCS()))
	d, err := dclient.Get(m.getTaskObjectName())

	util.SetNestedField(m.templateValues, d, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getCoreV1PVC will get the PVC as specified in the RunTask
func (m *taskExecutor) getCoreV1PVC() (err error) {
	pvc, err := m.getK8sClient().GetCoreV1PVCAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, pvc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getCoreV1PV will get the PersistentVolume as specified in the RunTask
func (m *taskExecutor) getCoreV1PV() (err error) {
	pv, err := m.getK8sClient().GetCoreV1PersistentVolumeAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, pv, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// getBatchV1Job will get the Job as specified in the RunTask
func (m *taskExecutor) getBatchV1Job() (err error) {
	job, err := m.getK8sClient().GetBatchV1JobAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}
	util.SetNestedField(m.templateValues, job, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// deleteBatchV1Job will delete one or more Jobs specified in the RunTask
func (m *taskExecutor) deleteBatchV1Job() (err error) {
	jobs := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")
	for _, name := range jobs {
		err = m.getK8sClient().DeleteBatchV1Job(strings.TrimSpace(name))
		if err != nil {
			return
		}
	}
	return
}

// deleteAppsV1STS will delete one or more StatefulSets
func (m *taskExecutor) deleteAppsV1STS() (err error) {
	stss := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")
	for _, name := range stss {
		err = m.getK8sClient().DeleteAppsV1STS(strings.TrimSpace(name))
		if err != nil {
			return
		}
	}
	return
}

// getStorageV1SC will get the StorageClass as specified in the RunTask
func (m *taskExecutor) getStorageV1SC() (err error) {
	sc, err := m.getK8sClient().GetStorageV1SCAsRaw(m.getTaskObjectName())
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, sc, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putStoragePool will put a CStorPool as defined in the task
func (m *taskExecutor) putStoragePool() (err error) {
	c, err := m.asStoragePool()
	if err != nil {
		return
	}

	storagePool, err := m.getK8sClient().CreateOEV1alpha1SPAsRaw(c)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, storagePool, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putCStorVolume will put a CStorVolume as defined in the task
func (m *taskExecutor) putCStorPool() (err error) {
	c, err := m.asCStorPool()
	if err != nil {
		return
	}

	cstorPool, err := m.getK8sClient().CreateOEV1alpha1CSPAsRaw(c)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, cstorPool, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putCStorVolume will put a CStorVolume as defined in the task
func (m *taskExecutor) putCStorVolume() (err error) {
	c, err := m.asCStorVolume()
	if err != nil {
		return
	}

	cstorVolume, err := m.getK8sClient().CreateOEV1alpha1CVAsRaw(c)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, cstorVolume, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// putCStorVolumeReplica will put a CStorVolumeReplica as defined in the task
func (m *taskExecutor) putCStorVolumeReplica() (err error) {
	d, err := m.asCstorVolumeReplica()
	if err != nil {
		return
	}

	cstorVolumeReplica, err := m.getK8sClient().CreateOEV1alpha1CVRAsRaw(d)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, cstorVolumeReplica, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// deleteOEV1alpha1SP will delete one or more StoragePool as specified in
// the RunTask
func (m *taskExecutor) deleteOEV1alpha1SP() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1SP(name)
		if err != nil {
			return
		}
	}

	return
}

// deleteOEV1alpha1CSP will delete one or more CStorPool as specified in
// the RunTask
func (m *taskExecutor) deleteOEV1alpha1CSP() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1CSP(name)
		if err != nil {
			return
		}
	}

	return
}

// deleteOEV1alpha1CSV will delete one or more CStorVolume as specified in
// the RunTask
func (m *taskExecutor) deleteOEV1alpha1CSV() (err error) {
	objectNames := strings.Split(strings.TrimSpace(m.getTaskObjectName()), ",")

	for _, name := range objectNames {
		err = m.getK8sClient().DeleteOEV1alpha1CSV(name)
		if err != nil {
			return
		}
	}

	return
}

// execCoreV1Pod runs given command remotely in given container of given pod
// and post stdout and and stderr in JsonResult. You can get it using -
// {{- jsonpath .JsonResult "{.Stdout}" | trim | saveAs "XXX" .TaskResult | noop -}}
func (m *taskExecutor) execCoreV1Pod() (err error) {
	podexecopts, err := podexec.WithTemplate("execCoreV1Pod", m.runtask.Spec.Task, m.templateValues).
		AsAPIPodExec()
	if err != nil {
		return
	}

	result, err := m.getK8sClient().ExecCoreV1Pod(m.getTaskObjectName(), podexecopts)
	if err != nil {
		return
	}

	util.SetNestedField(m.templateValues, result, string(v1alpha1.CurrentJSONResultTLP))
	return
}

// rolloutStatus generates rollout status of a given resource form it's object details
func (m *taskExecutor) rolloutStatus() (err error) {
	if m.metaTaskExec.isRolloutstatusExtnV1B1Deploy() {
		err = m.extnV1B1DeploymentRollOutStatus()
	} else if m.metaTaskExec.isRolloutstatusAppsV1Deploy() {
		err = m.appsV1DeploymentRollOutStatus()
	} else {
		err = fmt.Errorf("failed to get rollout status : meta task not supported: task details '%+v'", m.metaTaskExec.getTaskIdentity())
	}
	return
}

// listK8sResources will list resources as specified in the RunTask
func (m *taskExecutor) listK8sResources() (err error) {
	opts, err := m.metaTaskExec.getListOptions()
	if err != nil {
		return
	}

	var op []byte
	kc := m.getK8sClient()

	if m.metaTaskExec.isListCoreV1Pod() {
		op, err = kc.ListCoreV1PodAsRaw(opts)
	} else if m.metaTaskExec.isListCoreV1Service() {
		op, err = kc.ListCoreV1ServiceAsRaw(opts)
	} else if m.metaTaskExec.isListExtnV1B1Deploy() {
		op, err = kc.ListExtnV1B1DeploymentAsRaw(opts)
	} else if m.metaTaskExec.isListAppsV1B1Deploy() {
		op, err = kc.ListAppsV1B1DeploymentAsRaw(opts)
	} else if m.metaTaskExec.isListCoreV1PVC() {
		op, err = kc.ListCoreV1PVCAsRaw(opts)
	} else if m.metaTaskExec.isListCoreV1PV() {
		op, err = kc.ListCoreV1PVAsRaw(opts)
	} else if m.metaTaskExec.isListOEV1alpha1Disk() {
		op, err = kc.ListOEV1alpha1DiskRaw(opts)
	} else if m.metaTaskExec.isListOEV1alpha1SP() {
		op, err = kc.ListOEV1alpha1SPRaw(opts)
	} else if m.metaTaskExec.isListOEV1alpha1CSP() {
		op, err = kc.ListOEV1alpha1CSPRaw(opts)
	} else if m.metaTaskExec.isListOEV1alpha1CVR() {
		op, err = kc.ListOEV1alpha1CVRRaw(opts)
	} else if m.metaTaskExec.isListOEV1alpha1CV() {
		op, err = kc.ListOEV1alpha1CVRaw(opts)
	} else {
		err = fmt.Errorf("failed to list k8s resources: meta task not supported: task details '%+v'", m.metaTaskExec.getTaskIdentity())
	}

	if err != nil {
		return
	}

	// set the json doc result
	util.SetNestedField(m.templateValues, op, string(v1alpha1.CurrentJSONResultTLP))
	return
}

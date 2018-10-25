/*
Copyright 2018 The OpenEBS Authors.
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

package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"errors"

	"github.com/golang/glog"
	"github.com/openebs/CITF/common"
	strutil "github.com/openebs/CITF/utils/string"
	sysutil "github.com/openebs/CITF/utils/system"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	storage_v1 "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	negativeIndexErrorMessage = "container index can not be negative"
)

// GetAllNamespacesCoreV1NamespaceArray returns V1NamespaceList of all the namespaces.
// :return kubernetes.client.models.v1_namespace_list.V1NamespaceList: list of namespaces.
func (k8s K8S) GetAllNamespacesCoreV1NamespaceArray() ([]core_v1.Namespace, error) {
	nsList, err := k8s.Clientset.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	return nsList.Items, err
}

// GetAllNamespacesMap returns list of the names of all the namespaces.
// :return: map[string]core_v1.Namespace: map of namespaces where key is namespace name (str)
// and value is corresponding k8s.io/api/core/v1.Namespace object.
func (k8s K8S) GetAllNamespacesMap() (map[string]core_v1.Namespace, error) {
	namespacesList, err := k8s.GetAllNamespacesCoreV1NamespaceArray()
	if err != nil {
		return nil, err
	}

	namespaces := map[string]core_v1.Namespace{}
	for _, ns := range namespacesList {
		namespaces[ns.Name] = ns
	}
	return namespaces, nil
}

// GetPod returns the Pod object for given podName in the given namespace.
// :return: *kubernetes.client.models.v1_pod.V1Pod: Pointer to Pod objects.
func (k8s K8S) GetPod(namespace, podName string) (*core_v1.Pod, error) {
	podsClient := k8s.Clientset.CoreV1().Pods(namespace)
	return podsClient.Get(podName, meta_v1.GetOptions{})
}

// GetPods returns all the Pods object which has a prefix specified in its name in the given namespace.
// it tries to get the pods which match the criteria only once.
// NOTE: it counts pods which are not even in ContainerCreating state yet. Deal with them properly.
func (k8s K8S) GetPods(namespace, podNamePrefix string) ([]core_v1.Pod, error) {
	var thePods []core_v1.Pod

	// List pods
	pods, err := k8s.Clientset.CoreV1().Pods(namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return thePods, err
	}

	// Find the Pod
	logger.PrintlnDebugMessage(strings.Repeat("*", 80))
	logger.PrintfDebugMessage("all pods in %q namespace are:\n", namespace)
	for _, pod := range pods.Items {
		logger.PrintlnDebugMessage("complete Pod name is:", pod.Name)
		if strings.HasPrefix(pod.Name, podNamePrefix) {
			thePods = append(thePods, pod)
		}
	}
	logger.PrintlnDebugMessage(strings.Repeat("*", 80))

	return thePods, err
}

// GetPodsUntilQuitSignal returns all the Pods object which has a prefix specified in its name in the given namespace.
// it tries to get the pods which match the criteria unless `true` recieved from `quit` or it gets at least one such pod.
// NOTE: it counts pods which are not even in ContainerCreating state yet. Deal with them properly.
func (k8s K8S) GetPodsUntilQuitSignal(namespace, podNamePrefix string, quit <-chan bool) (thePods []core_v1.Pod, err error) {
	for {
		select {
		case quitting := <-quit:
			if quitting {
				if len(thePods) == 0 {
					err = fmt.Errorf("failed to get any pod which starts with %q, forced to quit", podNamePrefix)
				} else {
					glog.Info("quit signal recieved `true`, quitting...")
					err = nil
				}
				return
			}
			glog.Info("quit signal recieved `false`, not quitting...")

		default:
			thePods, err = k8s.GetPods(namespace, podNamePrefix)
			logger.LogErrorf(err, "error getting pods")
			if err == nil && len(thePods) != 0 {
				return thePods, nil
			}
			// If no pods found and debug is enabled then
			logger.PrintlnDebugMessage("no pods found in this iteration")

			time.Sleep(time.Second)
		}
	}
}

// GetPodsOrTimeout returns all the Pods object which has a prefix specified in its name in the given namespace.
// it tries to get the pods which match the criteria unless timeout occurs or it gets at least one such pod.
// NOTE: it counts pods which are not even in ContainerCreating state yet. Deal with them properly.
func (k8s K8S) GetPodsOrTimeout(namespace, podNamePrefix string, timeout time.Duration) ([]core_v1.Pod, error) {
	quit := make(chan bool)
	time.AfterFunc(timeout, func() {
		glog.Infof("timeout of duration %v ends", timeout)
		quit <- true
	})

	return k8s.GetPodsUntilQuitSignal(namespace, podNamePrefix, quit)
}

// GetPodsOrBlock returns all the Pods object which has a prefix specified in its name in the given namespace.
// it tries to get the pods which match the criteria unless it gets at least one such pod.
// NOTE: it counts pods which are not even in ContainerCreating state yet. Deal with them properly.
func (k8s K8S) GetPodsOrBlock(namespace, podNamePrefix string) ([]core_v1.Pod, error) {
	return k8s.GetPodsUntilQuitSignal(namespace, podNamePrefix, nil)
}

// ReloadPod reloads the state of the pod supplied and return error if any
func (k8s K8S) ReloadPod(pod *core_v1.Pod) (*core_v1.Pod, error) {
	return k8s.GetPod(pod.Namespace, pod.Name)
}

// GetPodPhase returns phase of the pod passed as an k8s.io/api/core/v1.PodPhase object.
//		:param k8s.io/api/core/v1.Pod pod: pod object for which you want to get phase.
//		:return: k8s.io/api/core/v1.PodPhase: phase of the pod.
func (k8s K8S) GetPodPhase(pod *core_v1.Pod) core_v1.PodPhase {
	return pod.Status.Phase
}

// GetPodPhaseStr returns phase of the pod passed in string format.
//		:param k8s.io/api/core/v1.Pod pod: pod object for which you want to get phase.
//		:return: str: phase of the pod.
func (k8s K8S) GetPodPhaseStr(pod *core_v1.Pod) string {
	return string(k8s.GetPodPhase(pod))
}

// GetContainerStatesInPod tries to get the states of all the containers of the supplied Pod only once
//    :param pod: pod object on which operation should be performed
//    :return: []k8s.io/api/core/v1.ContainerState: slice which holds states of the containers.
//           : error: error if occurred, `nil` otherwise
func (k8s K8S) GetContainerStatesInPod(pod *core_v1.Pod) (containerStates []core_v1.ContainerState, err error) {
	// Check if pointer to pod is nil
	if pod == nil {
		err = errors.New("nil argument supplied for pod")
		return
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerStates = append(containerStates, containerStatus.State)
	}
	return
}

// GetContainerStatesInPodUntilToldToQuit tries to get the states of all the containers of the supplied Pod
// until `true` is sent in `quit` channel
//    :param pod: pod object on which operation should be performed
//    :param quit: channel which is used to stop this function.
//    :return: []k8s.io/api/core/v1.ContainerState: slice which holds states of the containers.
//           : error: error if occurred, `nil` otherwise
func (k8s K8S) GetContainerStatesInPodUntilToldToQuit(pod *core_v1.Pod, quit <-chan bool) (containerStates []core_v1.ContainerState, err error) {
	for {
		select {
		case quitting := <-quit:
			if quitting {
				return
			}
		default:
			pod, err = k8s.ReloadPod(pod)
			if err != nil {
				continue
			}
			containerStates, err = k8s.GetContainerStatesInPod(pod)
			if err != nil || len(containerStates) == 0 {
				continue
			}
			return
		}
	}
}

// GetContainerStatesInPodWithTimeout returns the states of all the containers of the supplied Pod.
//    :param pod: pod object on which operation should be performed
//    :param timeout: maximum time duration to get the container's state.
//    :return: []k8s.io/api/core/v1.ContainerState: slice which holds states of the containers.
//           : error: error if occurred, `nil` otherwise
func (k8s K8S) GetContainerStatesInPodWithTimeout(pod *core_v1.Pod, timeout time.Duration) (containerStates []core_v1.ContainerState, err error) {
	quit := make(chan bool)
	done := make(chan bool)

	go func() {
		containerStates, err = k8s.GetContainerStatesInPodUntilToldToQuit(pod, quit)
		done <- true
	}()

	select {
	case <-time.After(timeout):
		quit <- true
	case <-done:
	}
	return
}

// GetContainerStateByIndexInPod tries to get the state of the container of supplied index of the supplied Pod only once
//    :param pod: pod object on which operation should be performed
//    :param containerIndex: index of the container for which you want state.
//    :return: k8s.io/api/core/v1.ContainerState: state of the container.
//           : error: error if occurred, `nil` otherwise
func (k8s K8S) GetContainerStateByIndexInPod(pod *core_v1.Pod, containerIndex int) (containerState core_v1.ContainerState, err error) {
	// check if container index is negative
	if containerIndex < 0 {
		err = errors.New(negativeIndexErrorMessage)
		return
	}

	// get state of all the containers in the pod supplied
	var containerStates []core_v1.ContainerState
	containerStates, err = k8s.GetContainerStatesInPod(pod)
	if err != nil {
		return
	}

	// if that pod has no containers
	if len(containerStates) == 0 {
		err = fmt.Errorf("no containers found in pod %q of namespace %q", pod.Name, pod.Namespace)
	} else if len(containerStates) < containerIndex { // if required number of container is not present
		err = fmt.Errorf("pod %q of namespace %q has only %d container(s) but expecting %d containers", pod.Name, pod.Namespace, len(containerStates), containerIndex+1)
		// inside this block expected number of containers (i. e. containerIndex+1) are always more than one
		// because control will enter this block only when number of containers is 1 or more than one
		// but when at least 1 container is present in the pod containerIndex can't be more or equal unless it is atleast 1
		// and when containerIndex is at least one then we are expecting at least 2 containers in the pod (number of containers = containerIndex+1)
		// thats why in above message I have written containers instead of container(s) for expected number.
	} else { // if there is enough containers
		containerState = containerStates[containerIndex]
	}
	return
}

// GetContainerStateByIndexInPodUntilToldToQuit tries to get the state of the container of supplied index of the supplied Pod
// until `true` is sent in `quit` channel
//    :param pod: pod object on which operation should be performed
//    :param containerIndex: index of the container for which you want state.
//    :param quit: channel which is used to stop this function.
//    :return: k8s.io/api/core/v1.ContainerState: state of the container.
//           : error: error if occurred, `nil` otherwise
func (k8s K8S) GetContainerStateByIndexInPodUntilToldToQuit(pod *core_v1.Pod, containerIndex int, quit <-chan bool) (containerState core_v1.ContainerState, err error) {
	for {
		select {
		case quitting := <-quit:
			if quitting {
				return
			}
		default:
			containerState, err = k8s.GetContainerStateByIndexInPod(pod, containerIndex)
			// If we get negative index then we should immediately return otherwise try again
			if err.Error() != negativeIndexErrorMessage && (err != nil || reflect.DeepEqual(containerState, core_v1.ContainerState{})) {
				continue
			}
			return
		}
	}
}

// GetContainerStateByIndexInPodWithTimeout returns the state of the container of supplied index of the supplied Pod.
//    :param pod: pod object on which operation should be performed
//    :param containerIndex: index of the container for which you want state.
//    :param timeout: maximum time duration to get the container's state.
//    :return: k8s.io/api/core/v1.ContainerState: state of the container.
//           : error: error if occurred, `nil` otherwise
func (k8s K8S) GetContainerStateByIndexInPodWithTimeout(pod *core_v1.Pod, containerIndex int, timeout time.Duration) (containerState core_v1.ContainerState, err error) {
	quit := make(chan bool)
	done := make(chan bool)

	go func() {
		containerState, err = k8s.GetContainerStateByIndexInPodUntilToldToQuit(pod, containerIndex, quit)
		done <- true
	}()

	select {
	case <-time.After(timeout):
		quit <- true
	case <-done:
	}
	return
}

// GetNodes returns a list of all the nodes.
//    :return: slice: list of nodes (slice of k8s.io/api/core/v1.Node array).
func (k8s K8S) GetNodes() (nodeNames []core_v1.Node, err error) {
	nodeNames = []core_v1.Node{}

	// To handle latency it tries 10 times each after 1 second of wait
	waited := 0
	for waited < 10 {
		nodeList, err := k8s.Clientset.CoreV1().Nodes().List(meta_v1.ListOptions{})
		if err != nil {
			break
		} else if len(nodeList.Items) == 0 {
			time.Sleep(time.Second)
			waited++
			continue
		}
		nodeNames = nodeList.Items
		break
	}

	return
}

// GetNodeNames returns a list of the name of all the nodes.
//    :return: slice: list of node names (slice of string array).
func (k8s K8S) GetNodeNames() (nodeNames []string, err error) {
	nodeNames = []string{}

	nodes, err := k8s.GetNodes()
	if err != nil {
		return
	}
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	return
}

// TODO: Write a function to label the node
// LabelNode label the node with the given key and value.
//    :param string node_name: Name of the node.
//    :param string key: Key of the label.
//    :param string value: Value of the label.
//    :return: error: if any error occurred or nil otherwise.
// func LabelNode(nodeName, key, value string) error { return fmt.Errorf("Not Implemented") }

// GetDaemonset returns the k8s.io/api/extensions/v1beta1.DaemonSet for the name supplied.
func (k8s K8S) GetDaemonset(daemonsetName, daemonsetNamespace string) (v1beta1.DaemonSet, error) {
	daemonsetClient := k8s.Clientset.ExtensionsV1beta1().DaemonSets(daemonsetNamespace)
	ds, err := daemonsetClient.Get(daemonsetName, meta_v1.GetOptions{})
	if err != nil {
		return v1beta1.DaemonSet{}, err
	}
	return *ds, nil
}

// ApplyDSFromManifestStruct Creates a Daemonset from the manifest supplied
func (k8s K8S) ApplyDSFromManifestStruct(manifest v1beta1.DaemonSet) (v1beta1.DaemonSet, error) {
	if manifest.Namespace == "" {
		manifest.Namespace = core_v1.NamespaceDefault
	}
	daemonsetClient := k8s.Clientset.ExtensionsV1beta1().DaemonSets(manifest.Namespace)
	ds, err := daemonsetClient.Create(&manifest)
	if err != nil {
		return v1beta1.DaemonSet{}, err
	}
	return *ds, nil
}

// GetDaemonsetStructFromYamlBytes returns k8s.io/api/extensions/v1beta1.DaemonSet
// for the yaml supplied
func (k8s K8S) GetDaemonsetStructFromYamlBytes(yamlBytes []byte) (v1beta1.DaemonSet, error) {
	ds := v1beta1.DaemonSet{}

	jsonBytes, err := strutil.ConvertYAMLtoJSON(yamlBytes)
	if err != nil {
		return ds, fmt.Errorf("error while Converting yaml string into Daemonset Structure. Error: %+v", err)
	}

	err = json.Unmarshal(jsonBytes, &ds)
	if err != nil {
		return ds, fmt.Errorf("error occurred while marshaling into Daemonset struct. Error: %+v", err)
	}

	return ds, nil
}

// TODO: Write a function to apply the YAML with the help of client-go
// YAMLApply apply the yaml specified by the argument.
//    :param str yamlPath: Path of the yaml file that is to be applied.
// func YAMLApplyAPI(yamlPath string) error { return fmt.Errorf("Not Implemented") }

// YAMLApply apply the yaml specified by the argument.
//    :param str yamlPath: Path of the yaml file that is to be applied.
func (k8s K8S) YAMLApply(yamlPath string) error {
	// TODO: Try using API call first. i.e. Using client-go

	err := sysutil.RunCommand(common.Kubectl + " apply -f " + yamlPath)
	logger.LogErrorf(err, "error occurred while applying the %s", yamlPath)
	if err != nil {
		return fmt.Errorf("failed applying %s", yamlPath)
	}
	return nil
}

// ExecToPodThroughAPI performs non-interactive exec to the pod with the specified command using client-go.
// :param string command: list of the str which specify the command.
// :param string containerName: name of the container in the Pod. (If the Pod has only one container, then it can be Empty String)
// :param string pod_name: Pod name
// :param string namespace: namespace of the Pod. (If it is blank string then, namespace will be default i.e. k8s.io/api/core/v1.NamespaceDefault)
// :param io.Reader stdin: Standard Input if necessary, otherwise `nil`
// :return: string: Output of the command. (STDOUT)
//          string: Errors. (STDERR)
//           error: If any error has occurred otherwise `nil`
func (k8s K8S) ExecToPodThroughAPI(command, containerName, podName, namespace string, stdin io.Reader) (string, string, error) {
	if len(namespace) == 0 {
		namespace = core_v1.NamespaceDefault
	}

	req := k8s.Clientset.Core().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	if err := core_v1.AddToScheme(scheme); err != nil {
		return "", "", fmt.Errorf("error adding to scheme: %v", err)
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	podExecOptions := core_v1.PodExecOptions{
		Command: strings.Fields(command),
		Stdin:   stdin != nil,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}
	if len(containerName) != 0 {
		podExecOptions.Container = containerName
	}

	req.VersionedParams(&podExecOptions, parameterCodec)

	logger.PrintlnDebugMessage("Request URL:", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(k8s.Config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", fmt.Errorf("error in Stream: %v", err)
	}

	return stdout.String(), stderr.String(), nil
}

// ExecToPodThroughKubectl performs non-interactive exec to the pod with the specified command using `kubectl exec`
// :param string command: list of the str which specify the command.
// :param string containerName: name of the container in the Pod. (If the Pod has only one container, then it can be Empty String)
// :param string pod_name: Pod name
// :param string namespace: namespace of the Pod. (If it is blank string then, namespace will be default i.e. k8s.io/api/core/v1.NamespaceDefault)
// :return: string: Output of the command. (STDOUT)
//           error: If any error has occurred otherwise `nil`
func (k8s K8S) ExecToPodThroughKubectl(command, containerName, podName, namespace string) (string, error) {
	kubectlCommand := common.Kubectl

	// adding namespace if namespace is not blank string
	if len(namespace) != 0 {
		kubectlCommand += " -n " + namespace
	}

	// adding podName
	kubectlCommand += " exec " + podName

	// adding container name if containerName is not a blank string
	if len(containerName) != 0 {
		kubectlCommand += " -c " + containerName
	}

	// finally adding command to execute
	kubectlCommand += " -- " + command

	return sysutil.ExecCommand(kubectlCommand)
}

// ExecToPod performs non-interactive exec to the pod with the specified command.
// first through API with `stdin` param as `nil`, if it fails then it uses `kubectl exec`
// :param string command: list of the str which specify the command.
// :param string containerName: name of the container in the Pod. (If the Pod has only one container, then it can be Empty String)
// :param string pod_name: Pod name
// :param string namespace: namespace of the Pod. (If it is blank string then, namespace will be default i.e. k8s.io/api/core/v1.NamespaceDefault)
// :return: string: Output of the command. (STDOUT)
//           error: If any error has occurred otherwise `nil`
func (k8s K8S) ExecToPod(command, containerName, podName, namespace string) (string, error) {
	stdout, stderr, err := k8s.ExecToPodThroughAPI(command, containerName, podName, namespace, nil)
	if err == nil {
		return stdout, nil
	}

	// When Exec through API fails
	glog.Errorf("error while exec into Pod through API. Stderr: %q. Error: %+v", stderr, err)
	return k8s.ExecToPodThroughKubectl(command, containerName, podName, namespace)
}

// GetLog returns the log of the pod.
// :param string pod_name: Name of the pod. (required)
// :param string namespace: Namespace of the pod. (required)
// :return: string: Log of the pod specified.
//           error: If an error has occurred, otherwise `nil`
func (k8s K8S) GetLog(podName, namespace string) (string, error) {
	// We can't declare a variable somewhere which can be skipped by goto
	var req *rest.Request
	var readCloser io.ReadCloser
	var err error

	buf := new(bytes.Buffer)
	req = k8s.Clientset.CoreV1().Pods(namespace).GetLogs(
		podName,
		&core_v1.PodLogOptions{},
	)

	readCloser, err = req.Stream()
	defer readCloser.Close()
	if err != nil {
		goto use_kubectl
	}

	buf.ReadFrom(readCloser)
	logger.PrintlnDebugMessage("Log of Pod", podName, "in Namespace", namespace, "through API:")
	logger.PrintlnDebugMessage(buf.String())

	return buf.String(), nil

use_kubectl:
	glog.Errorf("Error while getting log with API call. Error: %+v", err)

	return sysutil.ExecCommand(common.Kubectl + " -n " + namespace + " logs " + podName)
}

// BlockUntilPodIsUp blocks until all containers of the given pod is ready
// or when `true` is send to channel `quit`. It returns error if occurred.
func (k8s K8S) BlockUntilPodIsUp(pod *core_v1.Pod, quit <-chan bool) (err error) {
	var containerStates []core_v1.ContainerState
	var terminatedContainers int

	for {
		select {
		case quitting := <-quit:
			if quitting {
				glog.Info("forced to quit")
				return
			}
		default:
			if pod, err = k8s.ReloadPod(pod); err != nil {
				err = fmt.Errorf("error in reloading pod: %+v", err)
				goto continue_loop
			}

			containerStates, err = k8s.GetContainerStatesInPodUntilToldToQuit(pod, nil)
			logger.LogError(err, "error getting container states")

			// count terminated containers
			terminatedContainers = 0
			for _, containerState := range containerStates {
				if containerState.Terminated != nil {
					terminatedContainers++
				}
			}
			// if all containers are terminated return error
			if terminatedContainers == len(containerStates) {
				return fmt.Errorf("all containers in the pod %q of namespace %q have terminated", pod.Name, pod.Namespace)
			}

			// if any container is in waiting state
			for _, containerState := range containerStates {
				if containerState.Waiting != nil {
					if k8s.IsPodStateWait(containerState.Waiting.Reason) {
						fmt.Printf("waiting because pod-state: %q. Details: %+v\n", containerState.Waiting.Reason, *containerState.Waiting)
						goto continue_loop
					} else if !k8s.IsPodStateGood(containerState.Waiting.Reason) {
						return fmt.Errorf("pod %q of namespace %q is in bad state: %q. Details: %+v", pod.Name, pod.Namespace, containerState.Waiting.Reason, *containerState.Waiting)
					}
				}
			}

			for _, containerState := range containerStates {
				if containerState.Running == nil {
					// At this point all states are None,
					// so just showing phase is enough
					fmt.Printf("Waiting because pod %q of namespace %q is in phase: %q\n", pod.Name, pod.Namespace, k8s.GetPodPhase(pod))
					goto continue_loop
				}
			}

			goto break_loop
		}

	break_loop:
		break
	continue_loop:
		time.Sleep(time.Second)
		continue
	}

	return nil
}

// BlockUntilPodIsUpWithContext blocks until all containers of the given pod is ready
// or when supplied context cancelled. It returns error if occurred.
func (k8s K8S) BlockUntilPodIsUpWithContext(ctx context.Context, pod *core_v1.Pod) (err error) {
	quit := make(chan bool)
	done := make(chan error)
	go func() {
		done <- k8s.BlockUntilPodIsUp(pod, quit)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		quit <- true
		return fmt.Errorf("context cancelled while waiting for pod %q of namespace %q to be up", pod.Name, pod.Namespace)
	}
}

// BlockUntilPodIsUpOrTimeout blocks until all containers of the given pod is ready
// or when timeout is hit. It returns error if occurred.
// It uses `BlockUntilPodIsUpWithContext` internally, so in case of timeout it will give error describing "context cancelled"
func (k8s K8S) BlockUntilPodIsUpOrTimeout(pod *core_v1.Pod, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return k8s.BlockUntilPodIsUpWithContext(ctx, pod)
}

// GetDeployment returns the Deployment object for given deploymentName in the given namespace.
func (k8s K8S) GetDeployment(namespace, deploymentName string, opts meta_v1.GetOptions) (*v1beta1.Deployment, error) {
	deploymentClient := k8s.Clientset.ExtensionsV1beta1().Deployments(namespace)
	return deploymentClient.Get(deploymentName, opts)
}

// ListDeployments returns a pointer to the DeploymentList containing all the deployments.
func (k8s K8S) ListDeployments(namespace string, opts meta_v1.ListOptions) (*v1beta1.DeploymentList, error) {
	deploymentClient := k8s.Clientset.ExtensionsV1beta1().Deployments(namespace)
	return deploymentClient.List(opts)
}

// DeleteDeployment deletes the Deployment object of the given deploymentName in the given namespace.
func (k8s K8S) DeleteDeployment(namespace, deploymentName string, opts *meta_v1.DeleteOptions) error {
	deploymentClient := k8s.Clientset.ExtensionsV1beta1().Deployments(namespace)
	return deploymentClient.Delete(deploymentName, opts)
}

// GetStorageClass returns the StorageClass object for given storageClassName.
func (k8s K8S) GetStorageClass(storageClassName string, opts meta_v1.GetOptions) (*storage_v1.StorageClass, error) {
	storageClassClient := k8s.Clientset.StorageV1().StorageClasses()
	return storageClassClient.Get(storageClassName, opts)
}

// ListStorageClasses returns a pointer to StorageClassList containing all the storage classes.
func (k8s K8S) ListStorageClasses(opts meta_v1.ListOptions) (*storage_v1.StorageClassList, error) {
	storageClassClient := k8s.Clientset.StorageV1().StorageClasses()
	return storageClassClient.List(opts)
}

// DeleteStorageClass deletes the StorageClass object of given storageClassName.
func (k8s K8S) DeleteStorageClass(storageClassName string, opts *meta_v1.DeleteOptions) error {
	storageClassClient := k8s.Clientset.StorageV1().StorageClasses()
	return storageClassClient.Delete(storageClassName, opts)
}

// CreatePersistentVolumeClaim creates the PVC in the given namespace.
func (k8s K8S) CreatePersistentVolumeClaim(namespace string, persistentVolumeClaim *core_v1.PersistentVolumeClaim) (*core_v1.PersistentVolumeClaim, error) {
	persistentVolumeClaimClient := k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace)
	return persistentVolumeClaimClient.Create(persistentVolumeClaim)
}

// ListPersistentVolumeClaim lists all the PVCs in the given namespace.
func (k8s K8S) ListPersistentVolumeClaim(namespace string, opts meta_v1.ListOptions) (*core_v1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimClient := k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace)
	return persistentVolumeClaimClient.List(opts)
}

// GetPersistentVolumeClaim lists single PVC in the given namespace.
func (k8s K8S) GetPersistentVolumeClaim(namespace, persistentVolumeClaimName string, opts meta_v1.GetOptions) (*core_v1.PersistentVolumeClaim, error) {
	persistentVolumeClaimClient := k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace)
	return persistentVolumeClaimClient.Get(persistentVolumeClaimName, opts)
}

// DeletePersistentVolumeClaim deletes supplied PVC in the given namespace.
func (k8s K8S) DeletePersistentVolumeClaim(namespace, persistentVolumeClaimName string, opts *meta_v1.DeleteOptions) error {
	persistentVolumeClaimClient := k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace)
	return persistentVolumeClaimClient.Delete(persistentVolumeClaimName, opts)
}

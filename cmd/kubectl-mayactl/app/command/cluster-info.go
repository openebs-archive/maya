/*
Copyright 2019-20 The OpenEBS Authors.

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

package command

import (
	//"errors"

	//"html/template"
	//"strconv"
	//"strings"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clusterInfoCommandHelpText = `
This command fetches information and status of the various
aspects of the OPENEBS-control plane and add ons.

Usage: kubectl mayactl cluster-info
`
)

//ClusterComponentInfo keeps info of the control plane component of a current namespace
type ClusterComponentInfo struct {
	name   string
	ipaddr string
	status string
	mode   string
}

//NodeComponentInfo keeps the info for the  for the control plane node components of namespace
type NodeComponentInfo struct {
	name   string
	ipaddr string
	status string
}

//TODO:
//	-add volumeinfo as well for data plane components

// NewCmdClusterInfo displays OpenEBS Volume information.
func NewCmdClusterInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe",
		Aliases: []string{"cluster-info"},
		Short:   "Displays Openebs cluster info information",
		Long:    clusterInfoCommandHelpText,
		Example: `
		#To view the running control components of the cluster 
		$ mayactl cluster-info
		`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fetchComponentInfo())

		},
	}
	// TODO: allow for detection control plane elements ever in any other namespace but openebs
	//cmd.PersistentFlags().StringVarP(&options.namespace, "namespace", "n", options.namespace,
	//	"namespace name, required if volume is not in the default namespace")

	//TODO: add a flag for dump or flag[]
	return cmd
}

//checks the status of the cluster components that manage the OpenEBS compoenets
func fetchComponentInfo() (*[]ClusterComponentInfo, error) {

	clusterCompoenets := []ClusterComponentInfo{}

	//declare the selector to find the cluster components in the namespace
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement("app", selection.Exists, []string{})
	selector.Add(*requirement)

	//create the client to interact with the kube-api server
	kubeconfig := os.Getenv("KUBECONFIG")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := kubernetes.NewForConfig(config)

	// get a list of pods that have the label "openebs.io/conponent-name"
	pods, err := clientset.CoreV1().Pods("openebs").List(metav1.ListOptions{LabelSelector: selector.String()})

	if err != nil {
		panic(err.Error())
	}

	//print the components status at
	for _, pod := range pods.Items {
		component := ClusterComponentInfo{
			name:   pod.Status.ContainerStatuses[0].Name,
			ipaddr: pod.Status.PodIP,
			status: string(pod.Status.Phase),
			//mode:   pod.Status.mode,
		}

		clusterCompoenets = append(clusterCompoenets, component)
	}

	//separate the control plane node components & cluster components

	return &clusterCompoenets, nil
}

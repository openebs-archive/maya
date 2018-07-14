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
package pod

import (
	"fmt"
	"html/template"
	"os"

	internalk8sclient "github.com/openebs/maya/pkg/client/k8s"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
)

const podTemplate = `
================= Pod Details =====================
Name            		     Status
{{range $_, $value := .}}
{{$value.Name}}     {{$value.Status.Phase}}
{{end}}
===================================================
`

// CmdPodListOptions holds the options for pod list
type CmdPodListOptions struct {
	namespace    string
	allNameSpace bool
}

// NewCmdPodList lists all the pod, which running on openebs volumes
func NewCmdPodList() *cobra.Command {
	options := CmdPodListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the pods using openebs volumes",
		Run: func(_ *cobra.Command, _ []string) {
			err := options.listPod()
			if err != nil {
				fmt.Print(err)
			}
		},
	}
	cmd.Flags().StringVarP(&options.namespace, "namespace", "n", "default", "pod namespace.")
	cmd.Flags().BoolVarP(&options.allNameSpace, "all-namespaces", "a", false, "If present, list all the pods using openebs volumes across all namespaces.")
	return cmd
}

func (c *CmdPodListOptions) listPod() (err error) {
	namespace := c.namespace
	if c.allNameSpace {
		namespace = ""
	}
	clientSet, err := internalk8sclient.GetOutClusterCS()
	if err != nil {
		return
	}
	client := internalk8sclient.NewK8sClient(clientSet, nil, namespace)

	pods, err := client.GetPodWithEBSVolume()
	c.displayPods(pods)
	return
}

func (c *CmdPodListOptions) displayPods(pods []v1.Pod) {
	if len(pods) == 0 {
		fmt.Println("No Resources Found")
		return
	}
	tmpl, err := template.New("PodInfo").Parse(podTemplate)
	if err != nil {
		fmt.Println("Error displaying output, found error :", err)
		return
	}
	err = tmpl.Execute(os.Stdout, pods)
	if err != nil {
		fmt.Println("Error displaying pod details, found error :", err)
	}
}

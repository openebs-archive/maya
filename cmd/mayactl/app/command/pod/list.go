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

	internalk8sclient "github.com/openebs/maya/pkg/client/internalk8s"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdPodListOptions struct {
	namespace    string
	allNameSpace bool
}

func NewCmdPodList() *cobra.Command {
	options := CmdPodListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the pods running on openebs volumes",
		Run: func(_ *cobra.Command, _ []string) {
			err := options.ListPod()
			if err != nil {
				fmt.Print(err)
			}
		},
	}
	cmd.Flags().StringVarP(&options.namespace, "namespace", "n", "default", "pod namespace.")
	cmd.Flags().BoolVarP(&options.allNameSpace, "all-namespace", "a", false, "all the pod namespace.")
	return cmd
}

func (c *CmdPodListOptions) ListPod() (err error) {
	client, err := internalk8sclient.NewK8sClient()
	if err != nil {
		return
	}
	namespace := c.namespace
	if c.allNameSpace {
		namespace = ""
	}
	pods, err := client.GetPodWithEBSVolume(namespace)
	if len(pods) == 0 {
		fmt.Println("No Resources Found.")
		return
	}
	out := make([]string, len(pods)+1)
	var i int
	out[0] = "NAME|STATUS"
	for _, pod := range pods {
		i = i + 1
		out[i] = fmt.Sprintf("%s|%s", pod.GetName(), pod.Status.Phase)
	}
	fmt.Println(util.FormatList(out))
	return
}

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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	ndm "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
)

var (
	bdInfoCommandHelpText = `
This command lists the block devices
`
)

var namespace string

// NewCmdBlockDevices displays OpenEBS Volume information.
func NewCmdBlockDevices() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-devices",
		Aliases: []string{"block-device list"},
		Short:   "Displays Openebs bd list",
		Long:    bdInfoCommandHelpText,
		Example: `
		#To view the running control components of the cluster 
		$ mayactl cluster-info
		`,
		Run: func(cmd *cobra.Command, args []string) {
			listBlockDevices()

		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "openebs",
		"namespace name, required if volume is not in the default namespace")

	return cmd
}

//checks the status of the cluster components that manage the OpenEBS compoenets
func listBlockDevices() error {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, err := ndm.NewForConfig(config)

	if err != nil {

		fmt.Println(err)
		return err
	}

	bdList, err := clientset.OpenebsV1alpha1().BlockDevices(namespace).List(v1.ListOptions{})

	for _, bdObj := range bdList.Items {
		fmt.Println(bdObj.Name)
	}

	return nil
}

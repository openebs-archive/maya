/*
Copyright 2017 The OpenEBS Authors.

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

package pool

import (
	"fmt"
	"html/template"
	"os"
	"text/tabwriter"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

type pool struct {
	Name, Node, PoolType string
}

// CmdPoolOptions holds information of pool being operated
type CmdPoolOptions struct {
	poolName string
}

var (
	snapshotListCommandHelpText = `
This command displays available pools.

Usage: mayactl pool list

$ mayactl pool list
`

	options = &CmdPoolOptions{}
)

const poolListTemplate = `
{{ printf "%s\t" "POOL NAME"}} {{ printf "%s\t" "NODE NAME"}} {{ printf "%s\t" "POOL TYPE"}}
{{ printf "---------\t ---------\t ---------" }} {{range $key, $value := .}}
{{ printf "%v\t" $value.Name }} {{ printf "%v\t" $value.Node }} {{ printf "%v\t" $value.PoolType }} {{end}}
`

// NewCmdPoolList displays list of volumes
func NewCmdPoolList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all the pools",
		Long:  snapshotListCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunPoolList(cmd), util.Fatal)
		},
	}

	return cmd
}

// RunPoolList makes pool-list API request to maya-apiserver
func (c *CmdPoolOptions) RunPoolList(cmd *cobra.Command) error {
	resp, err := mapiserver.ListPools()
	if err != nil {
		return fmt.Errorf("Error listing pools: %v", err)
	}
	if len(resp.Items) == 0 {
		fmt.Println("No pools available")
		return nil
	}
	pools := make([]pool, 0)
	for _, p := range resp.Items {
		pools = append(pools, pool{
			Name:     p.GetName(),
			Node:     p.GetLabels()["kubernetes.io/hostname"],
			PoolType: p.Spec.PoolSpec.PoolType,
		})
	}
	print(poolListTemplate, pools)
	return nil
}

func print(format string, obj interface{}) error {
	// New Instance of tabwriter
	w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
	// New Instance of template
	tmpl, err := template.New("ReplicaStats").Parse(format)
	if err != nil {
		return fmt.Errorf("Error in parsing replica template, found error : %v", err)
	}
	// Parse struct with template
	err = tmpl.Execute(w, obj)
	if err != nil {
		return fmt.Errorf("Error in executing replica template, found error : %v", err)
	}
	return w.Flush()
}

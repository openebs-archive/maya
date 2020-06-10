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

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	poolDescribeCommandHelpText = `
This command displays available pools.

Usage: mayactl pool decribe -poolname <PoolName>

$ mayactl pool decribe -poolname <PoolName>
`
)

const poolDescribeTemplate = `
Pool Details :
--------------
Storage Pool Name  : {{ .ObjectMeta.Name }} 
Node Name          : {{ index .ObjectMeta.Labels "kubernetes.io/hostname" }}
CAS Template Used  : {{ index .ObjectMeta.Labels "openebs.io/cas-template-name" }}
CAS Type           : {{ index .ObjectMeta.Labels "openebs.io/cas-type" }}
StoragePoolClaim   : {{ index .ObjectMeta.Labels "openebs.io/storage-pool-claim" }}
UID                : {{ .ObjectMeta.UID }}
Pool Type          : {{ .Spec.PoolSpec.PoolType }}
Over Provisioning  : {{ .Spec.PoolSpec.OverProvisioning }}

Disk List :
-----------
{{ if eq (len .Spec.Group) 0 }}No disks present{{ else }}{{range $item := .Spec.Group }}{{range $disks := $item.Item }} {{printf "%s\n" $disks.Name }}{{ end }}{{ end }}{{ end }}
`

// NewCmdPoolDescribe displays info of pool
func NewCmdPoolDescribe() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describes the pools",
		Long:  poolDescribeCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.runPoolDescribe(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.poolName, "poolname", "", options.poolName,
		"a unique pool name.")
	return cmd
}

// runPoolDescrive makes pool-read API request to maya-apiserver
func (c *CmdPoolOptions) runPoolDescribe(cmd *cobra.Command) error {
	if len(c.poolName) == 0 {
		return fmt.Errorf("error: --poolname not specified")
	}
	resp, err := mapiserver.ReadPool(c.poolName)
	if err != nil {
		return fmt.Errorf("Error Reading pool: %v", err)
	}
	return mapiserver.Print(poolDescribeTemplate, resp)
}

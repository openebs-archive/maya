// Copyright © 2019-20 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCmdClusterInfo(t *testing.T) {
	tests := map[string]*struct {
		expectedCmd *cobra.Command
	}{
		"NewCmdClusterInfo": {
			expectedCmd: &cobra.Command{
				Use:     "cluster-info",
				Aliases: []string{"cluster-info"},
				Short:   "Displays Openebs cluster info information",
				Long:    clusterInfoCommandHelpText,
				Example: `#To view the running control components of the cluster 
				$ mayactl cluster-info
				`,
				Run: func(cmd *cobra.Command, args []string) {
					fetchComponentInfo()
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewCmdClusterInfo()
			if (got.Use != tt.expectedCmd.Use) || (reflect.DeepEqual(got.Aliases, tt.expectedCmd.Aliases) != true) || (got.Short != tt.expectedCmd.Short) ||
				(got.Long != tt.expectedCmd.Long) {
				t.Fatalf("TestName: %v | NewCmdPoolDescribe() => Got: %v | Want: %v \n", name, got, tt.expectedCmd)
			}
		})
	}

}

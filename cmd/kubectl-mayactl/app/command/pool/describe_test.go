// Copyright © 2017 The OpenEBS Authors
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

package pool

import (
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	utiltesting "k8s.io/client-go/util/testing"
)

func TestNewCmdPoolDescribe(t *testing.T) {
	tests := map[string]*struct {
		expectedCmd *cobra.Command
	}{
		"NewCmdVolumeDescribe": {
			expectedCmd: &cobra.Command{
				Use:   "describe",
				Short: "Describes the pools",
				Long:  poolDescribeCommandHelpText,
				Run: func(cmd *cobra.Command, args []string) {
					util.CheckErr(options.runPoolDescribe(cmd), util.Fatal)
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewCmdPoolDescribe()
			if (got.Use != tt.expectedCmd.Use) || (got.Short != tt.expectedCmd.Short) || (got.Long != tt.expectedCmd.Long) || (got.Example != tt.expectedCmd.Example) {
				t.Fatalf("TestName: %v | NewCmdPoolDescribe() => Got: %v | Want: %v \n", name, got, tt.expectedCmd)
			}
		})
	}
}

func TestRunPoolDescribe(t *testing.T) {
	options := CmdPoolOptions{}
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describes the pools",
		Long:  poolDescribeCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.runPoolDescribe(cmd), util.Fatal)
		},
	}
	tests := map[string]*struct {
		cmdPoolOptions *CmdPoolOptions
		cmd            *cobra.Command
		fakeHandler    utiltesting.FakeHandler
		err            error
		addr           string
	}{
		"StatusOK": {
			cmd: cmd,
			cmdPoolOptions: &CmdPoolOptions{
				poolName: "cstor-sparse-pool-qte1",
			},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: `{"metadata":{"creationTimestamp":"2018-11-16T11:08:49Z","generation":1,"labels":{"openebs.io/version":"0.8.0","kubernetes.io/hostname":"127.0.0.1","openebs.io/cas-template-name":"cstor-pool-create-default-0.8.0","openebs.io/cas-type":"cstor","openebs.io/cstor-pool":"cstor-sparse-pool-sst5","openebs.io/storage-pool-claim":"cstor-sparse-pool"},"name":"cstor-sparse-pool-sst5","resourceVersion":"584","selfLink":"/apis/openebs.io/v1alpha1/storagepools/cstor-sparse-pool-sst5","uid":"fea46be2-e98f-11e8-9c6a-b4b686bd0cff"},"spec":{"disks":{"diskList":["sparse-5a92ced3e2ee21eac7b930f670b5eab5"]},"format":"","message":"","mountpoint":"","name":"","nodename":"","path":"","poolSpec":{"cacheFile":"/tmp/cstor-sparse-pool.cache","overProvisioning":false,"poolType":"striped"}}}`,
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"Invalid Response": {
			cmd: cmd,
			cmdPoolOptions: &CmdPoolOptions{
				poolName: "cstor-sparse-pool-qte1",
			},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: `:"openebs.io/v1alpha1","kind":"StoragePool","metadata":{"labels":{"openebs.io/cas-template-name":"cstor-pool-create-default-0.8.0","openebs.io/cas-type":"cstor","openebs.io/cstor-pool":"cstor-sparse-pool-qte1","openebs.io/storage-pool-claim":"cstor-sparse-pool","kubernetes.io/hostname":"127.0.0.1"},"name":"cstor-sparse-pool-qte1","uid":"b5a62c11-e8eb-11e8-9ec2-b4b686bd0cff"},"spec":{"disks":{"diskList":["sparse-5a92ced3e2ee21eac7b930f670b5eab5"]},"format":"","message":"","mountpoint":"","name":"cstor-sparse-pool-qte1","nodename":"127.0.0.1","path":"","poolSpec":{"cacheFile":"","overProvisioning":false,"poolType":"striped"}}}`,
				T:            t,
			},
			err:  errors.New("Error Reading pool: invalid character ':' looking for beginning of value"),
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			cmd: cmd,
			cmdPoolOptions: &CmdPoolOptions{
				poolName: "cstor-sparse-pool-qte1",
			},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: "",
				T:            t,
			},
			err:  errors.New("Error Reading pool: Server status error: Not Found"),
			addr: "MAPI_ADDR",
		},
		"Response code 500": {
			cmd: cmd,
			cmdPoolOptions: &CmdPoolOptions{
				poolName: "cstor-sparse-pool-qte1",
			},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: "",
				T:            t,
			},
			err:  errors.New("Error Reading pool: Server status error: Internal Server Error"),
			addr: "MAPI_ADDR",
		},
		"When poolname is not specified": {
			cmd:            cmd,
			cmdPoolOptions: &CmdPoolOptions{},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: "",
				T:            t,
			},
			err:  errors.New("error: --poolname not specified"),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := tt.cmdPoolOptions.runPoolDescribe(tt.cmd)
			if !checkErr(got, tt.err) {
				t.Fatalf("TestName: %v | runPoolDescribe() => Got: %v | Want: %v \n", name, got, tt.err)
			}
		})
	}
}

package command

import (
	"testing"

	"net/http/httptest"
	"os"

	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	response1 = `{"metadata":{"annotations":{"vsm.openebs.io/targetportals":"<none>","vsm.openebs.io/cluster-ips":"<none>","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol","deployment.kubernetes.io/revision":"1","openebs.io/storage-pool":"default","vsm.openebs.io/replica-count":"1","openebs.io/jiva-controller-status":"Pending","openebs.io/volume-monitor":"false","openebs.io/replica-container-status":"Pending","openebs.io/jiva-controller-cluster-ip":"<none>","openebs.io/jiva-replica-status":"Pending","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol","openebs.io/capacity":"2G","openebs.io/jiva-controller-ips":"<none>","openebs.io/jiva-replica-ips":"<none>","vsm.openebs.io/replica-status":"Pending","vsm.openebs.io/controller-status":"Pending","openebs.io/controller-container-status":"Pending","vsm.openebs.io/replica-ips":"nil","openebs.io/jiva-target-portal":"nil","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"1","vsm.openebs.io/volume-size":"2G","vsm.openebs.io/controller-ips":""},"creationTimestamp":null,"labels":{},"name":"vol"},"status":{"Message":"","Phase":"Running","Reason":""}}`
)

func TestRunVolumeInfo(t *testing.T) {
	options := CmdVolumeInfoOptions{}
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Displays the info of Volume",
		Long:  volumeInfoCommandHelpText,

		Example: `mayactl volume info --volname <vol>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeInfo(cmd), util.Fatal)
		},
	}

	validCmd := map[string]struct {
		cmdOptions  *CmdVolumeInfoOptions
		cmd         *cobra.Command
		output      error
		err         error
		addr        string
		fakeHandler utiltesting.FakeHandler
	}{
		"WhenErrorGettingAnnotation": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			cmd: cmd,
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 200,
				//		ResponseBody: "",
				T: t,
			},
			addr:   "MAPI_ADDR",
			output: nil,
		},
		"WhenControllerIsNotRunning": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			cmd: cmd,
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(response1),
				T:            t,
			},
			addr:   "MAPI_ADDR",
			output: nil,
		},
	}
	for name, tt := range validCmd {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			if got := tt.cmdOptions.RunVolumeInfo(tt.cmd); got != tt.output {
				t.Fatalf("RunVolumeInfo(%v) => %v, want %v", tt.cmd, got, tt.output)
			}
			defer os.Unsetenv(tt.addr)
			defer server.Close()
		})
	}

}
func TestDisplayVolumeInfo(t *testing.T) {
	validInfo := map[string]struct {
		cmdOptions *CmdVolumeInfoOptions
		annotation *Annotations
		replica    client.Replica
		collection client.ReplicaCollection
		output     error
	}{
		"InfoWhenReplicaIsZero": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "0",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "",
			},
			output: nil,
		},
		"InfoWhenReplicaIsOne": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "1",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "10.10.10.10",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.10",
						Mode:    "RW",
					},
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsTwo": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "2",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running,Running",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "10.10.10.10,10.10.10.11",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.10",
						Mode:    "RW",
					},
					{
						Address: "10.10.10.11",
						Mode:    "RW",
					},
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndOnePending": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running,Running,Pending",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "10.10.10.10,10.10.10.11,nil",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.10",
						Mode:    "RW",
					},
					{
						Address: "10.10.10.11",
						Mode:    "RW",
					},
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndTwoPending": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running,Pending,Pending",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "10.10.10.10,nil,nil",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.10",
						Mode:    "RW",
					},
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndAllPending": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Pending,Pending,Pending",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "nil,nil,nil",
			},
			output: nil,
		},
		"InfoWhenReplicaIsThree": {
			cmdOptions: &CmdVolumeInfoOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running,Running,Running",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "10.10.10.10,10.10.10.11,10.10.10.12",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.10",
						Mode:    "RW",
					},
					{
						Address: "10.10.10.11",
						Mode:    "RW",
					},
					{
						Address: "10.10.10.12",
						Mode:    "RW",
					},
				},
			},
			output: nil,
		},
	}

	for name, tt := range validInfo {
		t.Run(name, func(t *testing.T) {
			if got := tt.cmdOptions.DisplayVolumeInfo(tt.annotation, tt.collection); got != tt.output {
				t.Fatalf("DisplayInfo(%v) => %v, want %v", tt.annotation, got, tt.output)
			}
		})
	}
}

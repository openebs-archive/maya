package command

import (
	"testing"

	"net/http/httptest"
	"os"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	response1 = `{"metadata":{"annotations":{"vsm.openebs.io/targetportals":"<none>","vsm.openebs.io/cluster-ips":"<none>","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol","deployment.kubernetes.io/revision":"1","openebs.io/storage-pool":"default","vsm.openebs.io/replica-count":"1","openebs.io/jiva-controller-status":"Pending","openebs.io/volume-monitor":"false","openebs.io/replica-container-status":"Pending","openebs.io/jiva-controller-cluster-ip":"<none>","openebs.io/jiva-replica-status":"Pending","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol","openebs.io/capacity":"2G","openebs.io/jiva-controller-ips":"<none>","openebs.io/jiva-replica-ips":"<none>","vsm.openebs.io/replica-status":"Pending","vsm.openebs.io/controller-status":"Pending","openebs.io/controller-container-status":"Pending","vsm.openebs.io/replica-ips":"nil","openebs.io/jiva-target-portal":"nil","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"1","vsm.openebs.io/volume-size":"2G","vsm.openebs.io/controller-ips":""},"creationTimestamp":null,"labels":{},"name":"vol"},"status":{"Message":"","Phase":"Running","Reason":""}}`
)

func TestRunVolumeInfo(t *testing.T) {
	options := CmdVolumeOptions{}
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Displays the info of Volume",
		Long:  volumeInfoCommandHelpText,

		Example: `mayactl volume info --volname <vol>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
			util.CheckErr(options.RunVolumeInfo(cmd), util.Fatal)
		},
	}

	validCmd := map[string]*struct {
		cmdOptions  *CmdVolumeOptions
		cmd         *cobra.Command
		output      error
		err         error
		addr        string
		fakeHandler utiltesting.FakeHandler
	}{
		"WhenErrorGettingAnnotation": {
			cmdOptions: &CmdVolumeOptions{
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
			cmdOptions: &CmdVolumeOptions{
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
		cmdOptions *CmdVolumeOptions
		annotation *Annotations
		replica    client.Replica
		collection client.ReplicaCollection
		volume     v1alpha1.CASVolume
		output     error
	}{
		"InfoWhenReplicaIsZero": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsOne": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsTwo": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndOnePending": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndTwoPending": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndAllPending": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThree": {
			cmdOptions: &CmdVolumeOptions{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAnd1stNodePendingo": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Pending,Running,Running",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "nil,10.10.10.11,10.10.10.12",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsTwoAndOneCrashLoopBackOff": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running,Running,CrashLoopBackOff",
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsThreeAndOneErrorPullBack": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "3",
				ControllerStatus: "Running",
				ReplicaStatus:    "Running,Running,ErrImagePull",
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
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsFourAndOneErrPullBackAndOneCrashBack": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "4",
				ControllerStatus: "Running",
				ReplicaStatus:    "Pending,ErrImagePull,Running,CrashLoopBackOff",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "nil,nil,10.10.10.12,nil",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.12",
						Mode:    "RW",
					},
				},
			},
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
		"InfoWhenReplicaIsFourAndOneErrPullBackAndOneCrashBackAndOneNil": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
			},
			annotation: &Annotations{
				TargetPortal:     "10.99.73.74:3260",
				ClusterIP:        "10.99.73.74",
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
				ReplicaCount:     "4",
				ControllerStatus: "Running",
				ReplicaStatus:    "Pending,ErrImagePull,Running,CrashLoopBackOff",
				VolSize:          "1G",
				ControllerIP:     "",
				Replicas:         "nil,nil,10.10.10.13,nil",
			},
			collection: client.ReplicaCollection{
				Data: []client.Replica{
					{
						Address: "10.10.10.13",
						Mode:    "RW",
					},
				},
			},
			volume: v1alpha1.CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-ips":    "10.48.1.17",
						"vsm.openebs.io/iqn":               "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/cluster-ips":           "10.51.242.184",
						"openebs.io/iqn":                   "iqn.2016-09.com.openebs.jiva:default-testclaimjiva",
						"openebs.io/replica-status":        "running, running, running",
						"vsm.openebs.io/cluster-ips":       "10.51.242.184",
						"vsm.openebs.io/replica-status":    "running, running, running",
						"vsm.openebs.io/volume-size":       "5G",
						"openebs.io/controller-ips":        "10.48.1.17",
						"openebs.io/volume-size":           "5G",
						"vsm.openebs.io/replica-count":     "3",
						"vsm.openebs.io/replica-ips":       "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/controller-status":     "running,running",
						"openebs.io/replica-count":         "3",
						"vsm.openebs.io/controller-status": "running,running",
						"openebs.io/replica-ips":           "10.48.0.7, 10.48.1.18, 10.48.2.7",
						"openebs.io/targetportals":         "10.51.242.184:3260",
						"vsm.openebs.io/targetportals":     "10.51.242.184:3260",
					},
				},
				Spec: v1alpha1.CASVolumeSpec{
					Capacity:     "5G",
					CasType:      "jiva",
					Iqn:          "iqn.2016-09.com.openebs.jiva:<no value>",
					Replicas:     "1",
					TargetIP:     "<no value>",
					TargetPort:   "3260",
					TargetPortal: "<no value>:3260",
				},
			},
			output: nil,
		},
	}

	for name, tt := range validInfo {
		t.Run(name, func(t *testing.T) {
			if got := tt.cmdOptions.DisplayVolumeInfo(tt.volume, tt.annotation, tt.collection); got != tt.output {
				t.Fatalf("DisplayInfo(%v) => %v, want %v", tt.annotation, got, tt.output)
			}
		})
	}
}

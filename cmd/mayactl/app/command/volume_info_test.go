package command

// var (
// 	response1 = `{"metadata":{"annotations":{"vsm.openebs.io/targetportals":"<none>","vsm.openebs.io/cluster-ips":"<none>","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol","deployment.kubernetes.io/revision":"1","openebs.io/storage-pool":"default","vsm.openebs.io/replica-count":"1","openebs.io/jiva-controller-status":"Pending","openebs.io/volume-monitor":"false","openebs.io/replica-container-status":"Pending","openebs.io/jiva-controller-cluster-ip":"<none>","openebs.io/jiva-replica-status":"Pending","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol","openebs.io/capacity":"2G","openebs.io/jiva-controller-ips":"<none>","openebs.io/jiva-replica-ips":"<none>","vsm.openebs.io/replica-status":"Pending","vsm.openebs.io/controller-status":"Pending","openebs.io/controller-container-status":"Pending","vsm.openebs.io/replica-ips":"nil","openebs.io/jiva-target-portal":"nil","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"1","vsm.openebs.io/volume-size":"2G","vsm.openebs.io/controller-ips":""},"creationTimestamp":null,"labels":{},"name":"vol"},"status":{"Message":"","Phase":"Running","Reason":""}}`
// )

// func TestRunVolumeInfo(t *testing.T) {
// 	options := CmdVolumeOptions{}
// 	cmd := &cobra.Command{
// 		Use:   "info",
// 		Short: "Displays the info of Volume",
// 		Long:  volumeInfoCommandHelpText,

// 		Example: `mayactl volume info --volname <vol>`,
// 		Run: func(cmd *cobra.Command, args []string) {
// 			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
// 			util.CheckErr(options.RunVolumeInfo(cmd), util.Fatal)
// 		},
// 	}

// 	validCmd := map[string]*struct {
// 		cmdOptions  *CmdVolumeOptions
// 		cmd         *cobra.Command
// 		output      error
// 		err         error
// 		addr        string
// 		fakeHandler utiltesting.FakeHandler
// 	}{
// 		"WhenErrorGettingAnnotation": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode: 200,
// 				//		ResponseBody: "",
// 				T: t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 		"When response code is 500": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode: 500,
// 				//		ResponseBody: "",
// 				T: t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 		"When response code is 404": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode: 404,
// 				//		ResponseBody: "",
// 				T: t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 		"When response code is 503": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode: 503,
// 				//		ResponseBody: "",
// 				T: t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 		"When response code is 600": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode: 600,
// 				//		ResponseBody: "",
// 				T: t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 		"When status in pending": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode:   200,
// 				ResponseBody: `{"metadata":{"annotations":{"vsm.openebs.io/targetportals":"<none>","vsm.openebs.io/cluster-ips":"<none>","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol","deployment.kubernetes.io/revision":"1","openebs.io/storage-pool":"default","vsm.openebs.io/replica-count":"1","openebs.io/jiva-controller-status":"Pending","openebs.io/volume-monitor":"false","openebs.io/replica-container-status":"Pending","openebs.io/jiva-controller-cluster-ip":"<none>","openebs.io/jiva-replica-status":"Pending","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol","openebs.io/capacity":"2G","openebs.io/jiva-controller-ips":"<none>","openebs.io/jiva-replica-ips":"<none>","vsm.openebs.io/replica-status":"Pending","vsm.openebs.io/controller-status":"Pending","openebs.io/controller-container-status":"Pending","vsm.openebs.io/replica-ips":"nil","openebs.io/jiva-target-portal":"nil","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"1","vsm.openebs.io/volume-size":"2G","vsm.openebs.io/controller-ips":""},"creationTimestamp":null,"labels":{},"name":"vol"},"status":{"Message":"","Phase":"pending","Reason":"pending"}}`,
// 				T:            t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 		"WhenControllerIsNotRunning": {
// 			cmdOptions: &CmdVolumeOptions{
// 				volName: "vol1",
// 			},
// 			cmd: cmd,
// 			fakeHandler: utiltesting.FakeHandler{
// 				StatusCode:   200,
// 				ResponseBody: string(response1),
// 				T:            t,
// 			},
// 			addr:   "MAPI_ADDR",
// 			output: nil,
// 		},
// 	}
// 	for name, tt := range validCmd {
// 		t.Run(name, func(t *testing.T) {
// 			server := httptest.NewServer(&tt.fakeHandler)
// 			os.Setenv(tt.addr, server.URL)
// 			if got := tt.cmdOptions.RunVolumeInfo(tt.cmd); got != tt.output {
// 				t.Fatalf("RunVolumeInfo(%v) => %v, want %v", tt.cmd, got, tt.output)
// 			}
// 			defer os.Unsetenv(tt.addr)
// 			defer server.Close()
// 		})
// 	}

// }

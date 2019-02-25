package command

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	utiltesting "k8s.io/client-go/util/testing"
)

func TestValidateResize(t *testing.T) {
	options := CmdVolumeOptions{}
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize the cStor Volume",
		Long:  volumeInfoCommandHelpText,

		Example: `mayactl volume info --volname <vol>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.ValidateResize(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeResize(cmd), util.Fatal)
		},
	}

	validCmd := map[string]*struct {
		cmdOptions     *CmdVolumeOptions
		cmd            *cobra.Command
		expectedOutput error
	}{
		"When volume size is missed": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
			},
			cmd:            cmd,
			expectedOutput: fmt.Errorf("--size is missing. Please specify value"),
		},
		"When volume name is missed": {
			cmdOptions: &CmdVolumeOptions{
				size: "5G",
			},
			cmd:            cmd,
			expectedOutput: fmt.Errorf("--volname is missing. Please specify the name of the volume to be resized"),
		},
		"When invalid size is given": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol2",
				size:    "-4OG",
			},
			cmd:            cmd,
			expectedOutput: fmt.Errorf("Please provide valid size and unit"),
		},
		"When invalid size unit is given": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol3",
				size:    "40TiB",
			},
			cmd:            cmd,
			expectedOutput: fmt.Errorf("Please provide valid size and unit"),
		},
		"When valid arguments are passed": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol3",
				size:    "40Pi",
			},
			cmd:            cmd,
			expectedOutput: nil,
		},
	}
	for name, tt := range validCmd {
		t.Run(name, func(t *testing.T) {
			err := tt.cmdOptions.ValidateResize(tt.cmd)
			if (err != nil && tt.expectedOutput != nil) && string(err.Error()) != string(tt.expectedOutput.Error()) {
				t.Errorf("Expected Output was %v \nbut got: %v", tt.expectedOutput, err)
			}
		})
	}
}

func TestRunVolumeResize(t *testing.T) {
	options := CmdVolumeOptions{}
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize the cStor Volume",
		Long:  volumeInfoCommandHelpText,

		Example: `mayactl volume info --volname <vol>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.ValidateResize(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeResize(cmd), util.Fatal)
		},
	}
	tests := map[string]*struct {
		cmdOptions     *CmdVolumeOptions
		cmd            *cobra.Command
		expectedOutput error
		addr           string
		fakeHandler    utiltesting.FakeHandler
	}{
		"Getting status error 400": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol1",
				size:    "40Pi",
			},
			cmd: cmd,
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string("HTTP Error 400 : Bad Request"),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: fmt.Errorf("Volume resize failed: HTTP Error 400 : Bad Request"),
		},
		"Getting status error 404": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol2",
				size:    "40Zi",
			},
			cmd: cmd,
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: string("HTTP Error 404 : Not Found"),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: fmt.Errorf("Volume resize failed: HTTP Error 404 : Not Found"),
		},
		"Resizing the volume": {
			cmdOptions: &CmdVolumeOptions{
				volName: "vol3",
				size:    "40Z",
			},
			cmd: cmd,
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(""),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			os.Setenv(tt.addr, server.URL)
			err := tt.cmdOptions.RunVolumeResize(cmd)
			if (err != nil && tt.expectedOutput != nil) && string(err.Error()) != string(tt.expectedOutput.Error()) {
				t.Errorf("\nExpected output was : %v \nbut got : %v", tt.expectedOutput, err)
			} else if (err != nil && tt.expectedOutput == nil) || (err == nil && tt.expectedOutput != nil) {
				t.Errorf("\nExpected output was : %v \nbut got : %v", tt.expectedOutput, err)
			}
		})
	}
}

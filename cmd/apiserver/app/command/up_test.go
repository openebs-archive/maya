package command

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func TestCommand_Implements(t *testing.T) {
	var _ cli.Command = &UpCommand{}
}

func TestCommand_Args(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "mayaserver")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	type tcase struct {
		args   []string
		errOut string
	}
	tcases := []tcase{
		{
			[]string{},
			"",
		},
		{
			[]string{"-data-dir=" + tmpDir},
			"",
		},
		{
			[]string{"-region=BANG-EAST"},
			"",
		},
	}
	for _, tc := range tcases {
		// Make a new command. We pre-emptively close the shutdownCh
		// so that the command exits immediately instead of blocking.
		ui := new(cli.MockUi)
		shutdownCh := make(chan struct{})
		close(shutdownCh)
		cmd := &UpCommand{
			Ui:         ui,
			ShutdownCh: shutdownCh,
		}

		// To prevent test failures on hosts whose hostname resolves to
		// a loopback address, we must append a bind address
		tc.args = append(tc.args, "-bind=169.254.0.1")
		if code := cmd.Run(tc.args); code != 1 {
			t.Fatalf("args: %v\nexit: %d\n", tc.args, code)
		}

		if expect := tc.errOut; expect != "" {
			out := ui.ErrorWriter.String()
			if !strings.Contains(out, expect) {
				t.Fatalf("expect to find %q\n\n%s", expect, out)
			}
		}
	}
}

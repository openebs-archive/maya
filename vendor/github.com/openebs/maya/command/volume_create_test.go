package command

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func TestVsmCreateCommand_Implements(t *testing.T) {
	var _ cli.Command = &VsmCreateCommand{}
}

func TestVsmCreateCommand_With_Meta(t *testing.T) {

	ui := new(cli.MockUi)

	cmd := &VsmCreateCommand{
		Meta: Meta{Ui: ui},
		Cmd:  exec.Command(string(ExecTesting), []string{""}...),
	}

	fh, err := ioutil.TempFile("", "maya")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh.Name())

	_, err = fh.WriteString(`
job "job1" {
	type = "service"
	datacenters = [ "dc1" ]
	group "group1" {
		count = 1
		task "task1" {
			driver = "exec"
			resources = {
				cpu = 1000
				memory = 512
			}
		}
	}
}`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if code := cmd.Run([]string{fh.Name()}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expected 'error starting cmd', got: %s", out)
	}
}

func TestVsmCreateCommand_Negative(t *testing.T) {

	ui := new(cli.MockUi)

	cmd := &VsmCreateCommand{
		Meta: Meta{Ui: ui},
		Cmd:  exec.Command(string(ExecTesting), []string{""}...),
	}

	// Fails on misuse
	if code := cmd.Run([]string{"some", "bad", "args"}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}
	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Usage: maya volume create") {
		t.Fatalf("expected 'usage: maya volume create', got: %s", out)
	}
	ui.ErrorWriter.Reset()

	// Fails when specified file does not exist
	cmd.Cmd = exec.Command(string(ExecTesting), []string{""}...)
	if code := cmd.Run([]string{"/unicorns/leprechauns"}); code != 1 {
		t.Fatalf("expect exit 1, got: %d", code)
	}
	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expect 'error starting cmd', got: %s", out)
	}
	ui.ErrorWriter.Reset()

	// Fails on invalid HCL
	fh1, err := ioutil.TempFile("", "maya")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh1.Name())

	if _, err := fh1.WriteString("nope"); err != nil {
		t.Fatalf("err: %s", err)
	}

	cmd.Cmd = exec.Command(string(ExecTesting), []string{""}...)
	if code := cmd.Run([]string{fh1.Name()}); code != 1 {
		t.Fatalf("expect exit 1, got: %d", code)
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expect 'error starting cmd', got: %s", out)
	}

	ui.ErrorWriter.Reset()

	// Fails on invalid job spec
	fh2, err := ioutil.TempFile("", "maya")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh2.Name())

	if _, err := fh2.WriteString(`job "job1" {}`); err != nil {
		t.Fatalf("err: %s", err)
	}

	cmd.Cmd = exec.Command(string(ExecTesting), []string{""}...)
	if code := cmd.Run([]string{fh2.Name()}); code != 1 {
		t.Fatalf("expect exit 1, got: %d", code)
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expect 'error starting cmd', got: %s", out)
	}

	ui.ErrorWriter.Reset()

	// Fails on connection failure (requires a valid job)
	fh3, err := ioutil.TempFile("", "maya")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh3.Name())

	_, err = fh3.WriteString(`
job "job1" {
	type = "service"
	datacenters = [ "dc1" ]
	group "group1" {
		count = 1
		task "task1" {
			driver = "exec"
			resources = {
				cpu = 1000
				memory = 512
			}
		}
	}
}`)

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cmd.Cmd = exec.Command(string(ExecTesting), []string{""}...)
	if code := cmd.Run([]string{"-address=nope", fh3.Name()}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expected 'error starting vsm', got: %s", out)
	}

	// Fails on invalid check-index (requires a valid job)
	cmd.Cmd = exec.Command(string(ExecTesting), []string{""}...)
	if code := cmd.Run([]string{"-check-index=bad", fh3.Name()}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expected 'error starting vsm', got: %s", out)
	}

	ui.ErrorWriter.Reset()

}

func TestVsmCreateCommand_From_STDIN(t *testing.T) {
	_, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := new(cli.MockUi)

	cmd := &VsmCreateCommand{
		Meta: Meta{Ui: ui},
		Cmd:  exec.Command(string(ExecTesting), []string{""}...),
	}

	go func() {
		stdinW.WriteString(`
job "job1" {
  type = "service"
  datacenters = [ "dc1" ]
  group "group1" {
		count = 1
		task "task1" {
			driver = "exec"
			resources = {
				cpu = 1000
				memory = 512
			}
		}
	}
}`)
		stdinW.Close()
	}()

	args := []string{"-"}
	if code := cmd.Run(args); code != 1 {
		t.Fatalf("expected exit code 1, got %d: %q", code, ui.ErrorWriter.String())
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expected 'error starting vsm', got: %s", out)
	}
	ui.ErrorWriter.Reset()
}

func TestVsmCreateCommand_From_URL(t *testing.T) {

	ui := new(cli.MockUi)

	cmd := &VsmCreateCommand{
		Meta: Meta{Ui: ui},
		Cmd:  exec.Command(string(ExecTesting), []string{""}...),
	}

	args := []string{"https://example.com/foo/bar"}
	if code := cmd.Run(args); code != 1 {
		t.Fatalf("expected exit code 1, got %d: %q", code, ui.ErrorWriter.String())
	}

	if out := ui.ErrorWriter.String(); !strings.Contains(out, "Error starting cmd") {
		t.Fatalf("expected 'error starting vsm', got: %s", out)
	}
}

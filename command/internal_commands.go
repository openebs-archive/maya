package command

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/mitchellh/cli"
)

// Main command that Maya will use internally
type ExecCmdType string

// Sub command that will be used with above main command
type SubCmdType string

const (
	// Nomad is currently the underlying lib used by Maya
	ExecNomad ExecCmdType = "nomad"

	// Kubernetes may be another lib that can be used by Maya
	ExecKube ExecCmdType = "kubectl"

	// Will be used for unit testing purposes
	ExecTesting ExecCmdType = "nopes"
)

// Nomad specific commands
const (
	NomadRun    SubCmdType = "run"
	NomadStatus SubCmdType = "status"
)

var ErrMissingCommand error = errors.New("missing command")

type InternalCommand struct {
	Cmd *exec.Cmd
	Ui  cli.Ui
}

func (ic *InternalCommand) Execute() int {

	if ic.Cmd.Path == "" {
		ic.Ui.Error(fmt.Sprintf("Error: %s", ErrMissingCommand))
		return 1
	}

	// Capture the std err
	ic.Cmd.Stderr = os.Stderr

	// Pipe that is connected to the command's std output when the command
	// starts
	rdCloser, err := ic.Cmd.StdoutPipe()
	if nil != err {
		ic.Ui.Error(fmt.Sprintf("Error piping to cmd's std output: %s", err))
		return 1
	}

	// use a scanner to break into lines
	scanner := bufio.NewScanner(rdCloser)
	go func() {
		for scanner.Scan() {
			ic.Ui.Output(scanner.Text())
		}
	}()

	// start the command
	// It does not wait till completion
	if err := ic.Cmd.Start(); nil != err {
		ic.Ui.Error(fmt.Sprintf("Error starting cmd: %s", err))
		return 1
	}

	// It waits till the command exits
	// returns the exit code & releases associated resources
	if err = ic.Cmd.Wait(); nil != err {
		ic.Ui.Error(fmt.Sprintf("Error executing cmd: %s", err))
		return 1
	}

	return 0
}

package command

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/mitchellh/cli"
)

// Main command that Maya will use internally
type ExecCmdType string

// Sub command that will be used with above main command
type SubCmdType string

const (
	// Nomad is currently the underlying lib used by Maya
	// for VSM provisioning
	ExecNomad ExecCmdType = "nomad"

	// Kubernetes may be another lib that can be used
	// for VSM provisioning
	ExecKube ExecCmdType = "kubectl"

	// Will be used for unit testing purposes
	ExecTesting ExecCmdType = "nopes"
)

// Nomad specific commands
const (
	NomadRun    SubCmdType = "run"
	NomadStatus SubCmdType = "status"
	NomadPlan   SubCmdType = "plan"
)

// Install specific scripts, path etc
const (
	InstallScriptsPath string = "https://raw.githubusercontent.com/openebs/maya/master/scripts/"
	BootstrapFile      string = "install_bootstrap.sh"
	BootstrapFilePath  string = InstallScriptsPath + BootstrapFile
	MayaScriptsPath    string = "/etc/maya.d/"
	InstallConsul      string = MayaScriptsPath + "install_consul.sh"
)

var ErrMissingCommand error = errors.New("missing command")

type InternalCommand struct {
	Cmd *exec.Cmd
	Ui  cli.Ui
}

// This executes the provided OS command (`assuming the command
// to be available where this binary is running`). It returns
// 0 or 1 depending on successful or failure in execution.
// NOTE: It will return the exit code of the internal command if
// available
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

		// Capture the error code if any
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			return waitStatus.ExitStatus()
		}

		return 1
	}

	return 0
}

func execute(cmd *exec.Cmd, ui cli.Ui) int {

	ic := &InternalCommand{
		Cmd: cmd,
		Ui:  ui,
	}

	return ic.Execute()
}

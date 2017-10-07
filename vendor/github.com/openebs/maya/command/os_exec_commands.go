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
	InstallScriptsPath  string = "https://raw.githubusercontent.com/openebs/maya/master/scripts/"
	BootstrapScript     string = "install_bootstrap.sh"
	BootstrapScriptPath string = InstallScriptsPath + BootstrapScript
	MayaScriptsPath     string = "/etc/maya.d/scripts/"
	// Utility scripts
	GetPrivateIPScript string = MayaScriptsPath + "get_first_private_ip.sh"
	// Consul scripts
	InstallConsulScript     string = MayaScriptsPath + "install_consul.sh"
	SetConsulAsServerScript string = MayaScriptsPath + "set_consul_as_server.sh"
	SetConsulAsClientScript string = MayaScriptsPath + "set_consul_as_client.sh"
	StartConsulServerScript string = MayaScriptsPath + "start_consul_server.sh"
	StartConsulClientScript string = MayaScriptsPath + "start_consul_client.sh"
	// Nomad scripts
	InstallNomadScript     string = MayaScriptsPath + "install_nomad.sh"
	SetNomadAsServerScript string = MayaScriptsPath + "set_nomad_as_server.sh"
	SetNomadAsClientScript string = MayaScriptsPath + "set_nomad_as_client.sh"
	StartNomadServerScript string = MayaScriptsPath + "start_nomad_server.sh"
	StartNomadClientScript string = MayaScriptsPath + "start_nomad_client.sh"
	// Docker scripts
	InstallDockerScript string = MayaScriptsPath + "install_docker.sh"
	// Etcd scripts
	//InstallEtcdScript string = MayaScriptsPath + "install_etcd.sh"
	//StartEtcdScript   string = MayaScriptsPath + "start_etcd.sh"
	//SetEtcdScript     string = MayaScriptsPath + "set_etcd.sh"

	//MayaServer Scripts
	InstallMayaserverScript string = MayaScriptsPath + "install_mayaserver.sh"
	StartMayaServerScript   string = MayaScriptsPath + "start_mayaserver.sh"
	//flannel Scripts
	InstallFlannelScript string = MayaScriptsPath + "install_flannel.sh"
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
func (ic *InternalCommand) Execute(capturer ...*string) int {

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
			if len(capturer) > 0 {
				*capturer[0] = *capturer[0] + scanner.Text()
			}
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

func execute(cmd *exec.Cmd, ui cli.Ui, capturer ...*string) int {

	ic := &InternalCommand{
		Cmd: cmd,
		Ui:  ui,
	}

	return ic.Execute(capturer...)
}

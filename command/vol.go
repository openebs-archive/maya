package command

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ExecCmdType string

const (
	MayaExecNomadRun ExecCmdType = "nomad run"
	MayaExecTesting  ExecCmdType = "nopes"
)

type ExecCommand struct {
	Cmd ExecCmdType
}

type VolCommand struct {
	M    Meta
	Exec ExecCommand
}

func (c *VolCommand) Help() string {
	helpText := `
Usage: maya vol <path>

  Creates a new vol or updates an existing vol using
  the specification i.e. Nomad jobfile located at <path>.

  If the supplied path is "-", the jobfile is read from stdin. Otherwise
  it is read from the file at the supplied path or downloaded and
  read from URL specified.

  On successful vol submission and scheduling, exit code 0 will be
  returned. If there are job placement issues encountered
  (unsatisfiable constraints, resource exhaustion, etc), then the
  exit code will be 2. Any other errors, including client connection
  issues or internal errors, are indicated by exit code 1.

General Options:

  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *VolCommand) Synopsis() string {
	return "Create a new vol or update an existing vol"
}

func (c *VolCommand) Run(args []string) int {

	flags := c.M.FlagSet("vol", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

	//if err := flags.Parse(args); err != nil {
	//	return 1
	//}

	// Prepare the command
	cmd := exec.Command(string(c.Exec.Cmd), args...)

	// Capture the std err
	cmd.Stderr = os.Stderr

	// Pipe that is connected to the command's std output when the command
	// starts
	rdCloser, err := cmd.StdoutPipe()
	if nil != err {
		c.M.Ui.Error(fmt.Sprintf("Error piping to command's std output: %s", err))
		return 1
	}

	// use a scanner to break into lines
	scanner := bufio.NewScanner(rdCloser)
	go func() {
		for scanner.Scan() {
			c.M.Ui.Output(scanner.Text())
		}
	}()

	// start the command
	// It does not wait till completion
	if err := cmd.Start(); nil != err {
		c.M.Ui.Error(fmt.Sprintf("Error starting vol: %s", err))
		return 1
	}

	// It waits till the command exits
	// returns the exit code & releases associated resources
	if err = cmd.Wait(); nil != err {
		c.M.Ui.Error(fmt.Sprintf("Error executing vol: %s", err))
		return 1
	}

	return 0
}

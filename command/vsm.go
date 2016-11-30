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
	MayaExecNomadRun ExecCmdType = "nomad"
	MayaExecTesting  ExecCmdType = "nopes"
)

type ExecCommand struct {
	Cmd ExecCmdType
}

type VsmCommand struct {
	M    Meta
	Exec ExecCommand
}

func (c *VsmCommand) Help() string {
	helpText := `
Usage: maya vsm <path>

  Creates a new VSM or updates an existing VSM using
  the specification i.e. Nomad jobfile located at <path>.

  If the supplied path is "-", the jobfile is read from stdin. Otherwise
  it is read from the file at the supplied path or downloaded and
  read from URL specified.

  On successful vsm submission and scheduling, exit code 0 will be
  returned. If there are job placement issues encountered
  (unsatisfiable constraints, resource exhaustion, etc), then the
  exit code will be 2. Any other errors, including client connection
  issues or internal errors, are indicated by exit code 1.

General Options:

  ` + generalOptionsUsage() + `

Run Options:
  -check-index
    If set, the vsm is only registered or updated if the passed
    vsm modify index matches the server side version. If a check-index value of
    zero is passed, the vsm is only registered if it does not yet exist. If a
    non-zero value is passed, it ensures that the vsm is being updated from a
    known state. The use of this flag is most common in conjunction with plan
    command.
  -detach
    Return immediately instead of entering monitor mode. After vsm submission,
    the evaluation ID will be printed to the screen, which can be used to
    examine the evaluation using the eval-status command.
  -verbose
    Display full information.
  -vault-token
    If set, the passed Vault token is stored in the vsm before sending to the
    Nomad servers. This allows passing the Vault token without storing it in
    the vsm file. This overrides the token found in $VAULT_TOKEN environment
    variable and that found in the vsm.
  -output
    Output the JSON that would be submitted to the HTTP API without submitting
    the vsm.
`
	return strings.TrimSpace(helpText)
}

func (c *VsmCommand) Synopsis() string {
	return "Create a new VSM or update an existing VSM"
}

func (c *VsmCommand) Run(args []string) int {

  var detach, verbose, output bool
	var checkIndexStr, vaultToken string

  // Create the sub command along with common flags (`i.e. general options`)
	flags := c.Meta.FlagSet("vsm", FlagSetClient)
	
	// These are the flags that are understood by `nomad run`
  // These are passed through as-is (`these are also known as run options`)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.BoolVar(&detach, "detach", false, "")
	flags.BoolVar(&verbose, "verbose", false, "")
	flags.BoolVar(&output, "output", false, "")
	flags.StringVar(&checkIndexStr, "check-index", "", "")
	flags.StringVar(&vaultToken, "vault-token", "", "")

  // Set the help function
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

  // Validate the args that has been passed with the flags that were just 
  // set above
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got exactly one argument
	args = flags.Args()
	if len(args) != 1 {
		c.M.Ui.Error(c.Help())
		return 1
	}

  // This will execute the `run` command of Nomad
	subcmd := []string{"run"}
	args = append(subcmd, args...)

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
		c.M.Ui.Error(fmt.Sprintf("Error starting vsm: %s", err))
		return 1
	}

	// It waits till the command exits
	// returns the exit code & releases associated resources
	if err = cmd.Wait(); nil != err {
		c.M.Ui.Error(fmt.Sprintf("Error executing vsm: %s", err))
		return 1
	}

	return 0
}

package command

import (
	"os/exec"
	"strings"
)

type VsmCreateCommand struct {
	M   Meta
	Cmd *exec.Cmd
}

func (c *VsmCreateCommand) Help() string {
	helpText := `
Usage: maya vsm-create [options] <path>

  Creates a new VSM using the specification located at <path>.

  On successful vsm creation submission and scheduling, exit code 0 will be
  returned. If there are placement issues encountered
  (unsatisfiable constraints, resource exhaustion, etc), then the
  exit code will be 2. Any other errors, including client connection
  issues or internal errors, are indicated by exit code 1.

General Options:

  ` + generalOptionsUsage() + `

VSM Options:
  -verbose
    Display full information.
`
	return strings.TrimSpace(helpText)
}

func (c *VsmCreateCommand) Synopsis() string {
	return "Creates a new VSM"
}

func (c *VsmCreateCommand) Run(args []string) int {

	var verbose bool

	flags := c.M.FlagSet("vsm-create", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }
	flags.BoolVar(&verbose, "verbose", false, "")

	// Set the help function
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

	// Validate the args that has been passed against
	// the flags that were just defined above
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got exactly one argument
	args = flags.Args()
	if len(args) != 1 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	if c.Cmd == nil {
		// This will execute the `run` command of Nomad
		// subcmd := []string{string(NomadRun)}
		args = append([]string{string(NomadRun)}, args...)

		// Prepare the command
		c.Cmd = exec.Command(string(ExecNomad), args...)
	}

	ic := &InternalCommand{
		Cmd: c.Cmd,
		Ui:  c.M.Ui,
	}

	if op := ic.Execute(); 0 != op {
		c.M.Ui.Error("Error creating vsm")
		return 1
	}

	return 0
}

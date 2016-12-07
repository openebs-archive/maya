package command

import (
	"os/exec"
	"strings"
)

type VsmCreateCommand struct {
	// To control this CLI's display
	M Meta
	// OS command to execute; <optional>
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

VSM Create Options:
  -verbose
    Display full information.
`
	return strings.TrimSpace(helpText)
}

func (c *VsmCreateCommand) Synopsis() string {
	return "Creates a new VSM"
}

// The logic of this function can be understood by understanding
// the help text defined earlier.
func (c *VsmCreateCommand) Run(args []string) int {

	var verbose bool
	var op int

	flags := c.M.FlagSet("vsm-create", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }
	flags.BoolVar(&verbose, "verbose", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// specs file is mandatory
	args = flags.Args()
	if len(args) != 1 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	if c.Cmd == nil {
		// sub command
		args = append([]string{string(NomadRun)}, args...)

		// main command; append sub cmd to main cmd
		c.Cmd = exec.Command(string(ExecNomad), args...)
	}

	ic := &InternalCommand{
		Cmd: c.Cmd,
		Ui:  c.M.Ui,
	}

	if op = ic.Execute(); 0 != op {
		c.M.Ui.Error("Error creating vsm")
		return op
	}

	return op
}

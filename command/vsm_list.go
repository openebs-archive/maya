package command

import (
	"os/exec"
	"strings"
)

type VsmListCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute; <optional>
	Cmd *exec.Cmd

	// Check the help section to learn more on these variables
	length  int
	evals   bool
	verbose bool
}

func (c *VsmListCommand) Help() string {
	helpText := `
Usage: maya vsm-list [options] <vsm-id>

  Display status information about vsm(s). If no vsm ID is given,
  a list of all known vsms will be dumped. 
  
  NOTE: Provide a prefix of vsm ID if you have forgotten the entire ID.

General Options:

  ` + generalOptionsUsage() + `

VSM List Options:

  -short
    Display short output. Used only when a single vsm is being
    queried, and drops verbose information about allocations.

  -evals
    Display the evaluations associated with the vsm.

  -verbose
    Display full information.
`
	return strings.TrimSpace(helpText)
}

func (c *VsmListCommand) Synopsis() string {
	return "Display status information about vsm(s)"
}

func (c *VsmListCommand) Run(args []string) int {
	var short bool
	var op int

	flags := c.M.FlagSet("vsm-list", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }
	flags.BoolVar(&short, "short", false, "")
	flags.BoolVar(&c.evals, "evals", false, "")
	flags.BoolVar(&c.verbose, "verbose", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we either got no vsms or exactly one.
	args = flags.Args()
	if len(args) > 1 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	// TODO: Future might involve delegating to
	// Nomad or Kubectl based on some env property !!!
	// NOTE: args will not be used if Cmd is set previously
	if c.Cmd == nil {
		// sub command
		args = append([]string{string(NomadStatus)}, args...)

		// main command; append sub cmd to main cmd
		c.Cmd = exec.Command(string(ExecNomad), args...)
	}

	ic := &InternalCommand{
		Cmd: c.Cmd,
		Ui:  c.M.Ui,
	}

	if op = ic.Execute(); 0 != op {
		c.M.Ui.Error("Error listing vsm(s)")
		return op
	}

	return op
}

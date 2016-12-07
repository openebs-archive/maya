package command

import (
	"fmt"
	"os/exec"
	"strings"
)

type InstallMayaCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute; <optional>
	Cmd *exec.Cmd

	// Check the help section to learn more on these variables
	bootstrap bool
}

func (c *InstallMayaCommand) Help() string {
	helpText := `
Usage: maya install-maya

  Installs maya server.   

General Options:

  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *InstallMayaCommand) Synopsis() string {
	return "Installs maya server"
}

func (c *InstallMayaCommand) Run(args []string) int {
	var runop int

	flags := c.M.FlagSet("install-maya", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// There are no extra arguments
	oargs := flags.Args()
	if len(oargs) != 0 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	if c.Cmd != nil {
		// execute the provided command
		return execute(c.Cmd, c.M.Ui)
	}

	// install related steps
	args = append([]string{string(InstallBootstrapFile) + " | sh"}, oargs...)
	c.Cmd = exec.Command("curl -sSL ", args...)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Failed to bootstrap the install")
		c.M.Ui.Error(fmt.Sprintf("Verify presence of file: %s", InstallBootstrapFile))
		return runop
	}

	args = append([]string{string(MayaScriptsPath)}, oargs...)
	c.Cmd = exec.Command("ls", args...)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install bootstrap step is missing")
		return runop
	}

	args = append([]string{string(InstallConsul)}, oargs...)
	c.Cmd = exec.Command(string(ExecScript), args...)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Error installing consul")
		return runop
	}

	return runop
}

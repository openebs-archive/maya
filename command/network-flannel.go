package command

import (
	"os/exec"
	"strings"
)

type NetworkInstallCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute
	Cmd *exec.Cmd

	// self ip address
	self_ip string

	// self hostname
	self_hostname string
}

func (c *NetworkInstallCommand) Help() string {
	helpText := `
		Usage: maya network-install 
`
	return strings.TrimSpace(helpText)
}

func (c *NetworkInstallCommand) Synopsis() string {
	return "Configure flannel network on maya-master machine."

}
func (c *NetworkInstallCommand) Run(args []string) int {

	var runop int

	flags := c.M.FlagSet("network-install", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }
	flags.StringVar(&c.self_ip, "etcd-ip", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	//check the args
	args = flags.Args()
	if len(args) > 1 {
		c.M.Ui.Error(c.Help())
		return 1
	}
	if runop = c.installFlannel(); runop != 0 {
		return runop
	}
	return runop
}

func (c *NetworkInstallCommand) installFlannel() int {
	var runop int = 0

	c.Cmd = exec.Command("sh", InstallFlannelScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing flannel")
	}

	return runop
}

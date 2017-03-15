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
	kube_ip string

	// self hostname
	kubename string
}

func (c *NetworkInstallCommand) Help() string {
	helpText := `
	Usage: maya network-install <k8smaster-name> <ip-addr>

Maya Network options:
		

		`
	return strings.TrimSpace(helpText)
}

func (c *NetworkInstallCommand) Synopsis() string {
	return "Configure flannel network on maya-host machine."

}
func (c *NetworkInstallCommand) Run(args []string) int {

	var runop int

	flags := c.M.FlagSet("network-install", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }
	flags.StringVar(&c.kubename, "k8smaster-name", "", "")
	flags.StringVar(&c.kube_ip, "ip-addr", "", "")

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

	c.Cmd = exec.Command("sh", InstallFlannelScript, c.kube_ip, c.kubename)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing flannel")
	}

	return runop
}

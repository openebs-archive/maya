package command

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type NetworkInstallCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute
	Cmd *exec.Cmd

	// etcd ip address
	kube_ip string

	//Server name in which etcd is running
	kubename string

	// cni plugin-name to install as maya network
	cni string
}

func (c *NetworkInstallCommand) Help() string {
	helpText := `
	Usage: maya network-install <cni> <name> <ip>

	Configure the virtual network for containers on OpenEBS Host (osh)

Maya Network options:
  -cni= <Name>
    Name of the CNI plugin to configure as a virtual container network

  -name= <Name>
    This is name of the host which is running
    the etcd server to manage the key-value pair.
 
  -ip= <IP Address> 
    This args is ip-address of the same etcd server mentioned above 
	running on kubernetes-master.

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
	flags.StringVar(&c.kubename, "name", "", "")
	flags.StringVar(&c.kube_ip, "ip", "", "")
	flags.StringVar(&c.cni, "cni", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	//check the non-flag arguments
	args = flags.Args()
	if len(args) != 0 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	//check the args
	if len(strings.TrimSpace(c.cni)) == 0 {
		c.M.Ui.Error(fmt.Sprintf("-cni option is mandatory\n"))
		c.M.Ui.Error(c.Help())
		return 1
	}

	if len(strings.TrimSpace(c.kubename)) == 0 {
		c.M.Ui.Error(fmt.Sprintf("-name option is mandatory\n"))
		c.M.Ui.Error(c.Help())
		return 1
	}

	if len(strings.TrimSpace(c.kube_ip)) == 0 {
		c.M.Ui.Error(fmt.Sprintf("-ip option is mandatory\n"))
		c.M.Ui.Error(c.Help())
		return 1
	}

	//stdout the configuration
	fmt.Printf("following Configuration has been passed:\n")
	fmt.Printf("k8smaster-name = %v\n", c.kubename)
	fmt.Printf("k8smaster-ip = %v\n", c.kube_ip)
	fmt.Printf("cni-plugin = %v\n", c.cni)

	if runop = c.installFlannel(); runop != 0 {
		return runop
	}
	return runop
}

func (c *NetworkInstallCommand) installFlannel() int {
	var runop int = 0

	//Validation of ip
	var ipAddr net.IP
	if len(strings.TrimSpace(c.kube_ip)) > 0 {
		if ipAddr = net.ParseIP(c.kube_ip); ipAddr == nil {
			c.M.Ui.Error(fmt.Sprintf("provided ip address is not correct"))
			return 1
		}
	}

	c.Cmd = exec.Command("sh", InstallFlannelScript, c.kube_ip, c.kubename)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing flannel")
	}

	return runop
}

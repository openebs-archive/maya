package command

import (
	"os/exec"
	"strings"
)

type InstallOpenEBSCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute
	Cmd *exec.Cmd

	// all maya master ips, in a comma separated format
	master_ips string

	// self ip address
	self_ip string

	// all maya client ips, in a comma separated format
	member_ips string
}

func (c *InstallOpenEBSCommand) Help() string {
	helpText := `
Usage: maya setup-osh

  Configure this machine as OpenEBS Host and enable it 
  to run OpenEBS VSMs. 

General Options:

  ` + generalOptionsUsage() + `

OpenEBS Storage Host (osh) setup options:

  -omm-ips=<IP Address(es) of all maya masters>
    Comma separated list of IP addresses of all maya masters
    participating in the cluster.
    
  -self-ip=<IP Address>
    The IP Address of this local machine i.e. the machine where
    this command is being run. This is required when the machine
    has many private IPs and you want to use a specific IP.
  
    NOTE: Do not include the IP address of this local machine i.e.
    the machine where this command is being run.
`
	return strings.TrimSpace(helpText)
}

func (c *InstallOpenEBSCommand) Synopsis() string {
	return "Configure this machine as OpenEBS Host."
}

func (c *InstallOpenEBSCommand) Run(args []string) int {
	var runop int

	flags := c.M.FlagSet("setup-osh", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

	flags.StringVar(&c.master_ips, "omm-ips", "", "")
	flags.StringVar(&c.self_ip, "self-ip", "", "")
	flags.StringVar(&c.member_ips, "member-ips", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// There are no extra arguments
	oargs := flags.Args()
	if len(oargs) != 0 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	if len(strings.TrimSpace(c.master_ips)) == 0 {
		c.M.Ui.Error("-omm-ips option is mandatory")
		c.M.Ui.Error(c.Help())
		return 1
	}

	mi := &MayaAsNomadInstaller{
		InstallCommand: InstallCommand{
			Ui: c.M.Ui,
		},
		self_ip:    c.self_ip,
		client_ips: c.member_ips,
		master_ips: c.master_ips,
		is_master:  false,
	}

	if runop = mi.Install(); runop != 0 {
		c.M.Ui.Error("OpenEBS Host setup failed")
	}

	return runop
}

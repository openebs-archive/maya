package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/openebs/maya/pkg/client/mapiserver"
)

// VsmCreateCommand is a command implementation struct
type VsmCreateCommand struct {
	// To control this CLI's display
	Meta
	// OS command to execute; <optional>
	Cmd     *exec.Cmd
	vsmname string
	size    string
}

// Help shows helpText for a particular CLI command
func (c *VsmCreateCommand) Help() string {
	helpText := `
	Usage: maya volume create -volname <vol> [-size <size>]

	This command creates a new Volume.

	Volume create options:
	-size
	Provisioning size of the volume(default is 5G)

	`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmCreateCommand) Synopsis() string {
	return "Creates a new Volume"
}

// Run to get the flag values and start execution
//The logic of this function can be understood by understanding
// the help text defined earlier.
func (c *VsmCreateCommand) Run(args []string) int {

	var op int

	flags := c.Meta.FlagSet("volume create", FlagSetClient)
	flags.Usage = func() { c.Meta.Ui.Output(c.Help()) }
	flags.StringVar(&c.vsmname, "volname", "", "")
	flags.StringVar(&c.size, "size", "5G", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// specs file is mandatory
	args = flags.Args()
	if len(args) != 1 && len(strings.TrimSpace(c.vsmname)) == 0 {
		c.Ui.Error(c.Help())
		return 1
	}
	if len(args) == 1 {
		if c.Cmd == nil {
			// sub command
			args = append([]string{string(NomadRun)}, args...)

			// main command; append sub cmd to main cmd
			c.Cmd = exec.Command(string(ExecNomad), args...)
		}

		ic := &InternalCommand{
			Cmd: c.Cmd,
			Ui:  c.Ui,
		}

		if op = ic.Execute(); 0 != op {
			c.Ui.Error("Error creating Volume")
			return op
		}
		return 1
	}
	if c.vsmname != " " {
		if !strings.HasSuffix(c.size, "G") {
			fmt.Println("-size should contain the suffix 'G',which represent the size in GB (exp: 10G)")
			return 0
		}
		jobID := c.vsmname

		// Get the HTTP client
		client, err := c.Meta.Client()
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
			return 1
		}
		// Check if the VSM exists
		job, _, err := client.Jobs().Info(jobID, nil)
		if err == nil || job != nil {
			c.Ui.Error(fmt.Sprintf("Volume already exist: %q", jobID))
			return 1
		}

		resp := mapiserver.CreateVolume(c.vsmname, c.size)
		if resp != nil {
			c.Ui.Error(fmt.Sprintf("Error Creating Volume: %v", resp))
			return 1
		}
		fmt.Printf("Volume Successfully Created:%v\n", c.vsmname)
	}
	return op
}

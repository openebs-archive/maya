package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// VsmStopCommand is a command implementation struct
type VsmStopCommand struct {
	Meta
	volname string
}

// Help shows helpText for a particular CLI command
func (c *VsmStopCommand) Help() string {
	helpText := `
Usage: maya volume delete <vol>

This command deletes an existing volume.

`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmStopCommand) Synopsis() string {
	return "Deletes a running Volume"
}

// Run holds the flag values for CLI subcommands
func (c *VsmStopCommand) Run(args []string) int {
	var detach, verbose, autoYes bool

	flags := c.Meta.FlagSet("volume delete", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&c.volname, "volname", "", "Volume name")
	flags.BoolVar(&detach, "detach", false, "")
	flags.BoolVar(&verbose, "verbose", false, "")
	flags.BoolVar(&autoYes, "yes", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	addr := os.Getenv("KUBERNETES_SERVICE_HOST")
	if addr != "" {
		err := DeleteVsm(c.volname)
		if err != nil {
			fmt.Sprintf("Error while deleting Volume: %s", err)
		}
		return 0
	}

	// Truncate the id unless full length is requested
	length := shortId
	if verbose {
		length = fullId
	}

	// Check that we got exactly one vsm
	args = flags.Args()
	if len(args) != 1 {
		c.Ui.Error(c.Help())
		return 1
	}
	jobID := args[0]

	// Get the HTTP client
	client, err := c.Meta.Client()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	// Check if the VSM exists
	jobs, _, err := client.Jobs().PrefixList(jobID)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error deregistering Volume: %s", err))
		return 1
	}
	if len(jobs) == 0 {
		c.Ui.Error(fmt.Sprintf("No Volume(s) with prefix or id %q found", jobID))
		return 1
	}
	if len(jobs) > 1 && strings.TrimSpace(jobID) != jobs[0].ID {
		out := make([]string, len(jobs)+1)
		out[0] = "ID|Type|Priority|Status"
		for i, job := range jobs {
			out[i+1] = fmt.Sprintf("%s|%s|%d|%s",
				job.ID,
				job.Type,
				job.Priority,
				job.Status)
		}
		c.Ui.Output(fmt.Sprintf("Prefix matched multiple Volume(s)\n\n%s", formatList(out)))
		return 0
	}
	// Prefix lookup matched a single VSM
	job, _, err := client.Jobs().Info(jobs[0].ID, nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error deregistering job: %s", err))
		return 1
	}

	// Confirm the stop if the VSM was a prefix match.
	// to fix the --> pointers being printed while passing status commands
	if jobID != *job.ID && !autoYes {
		question := fmt.Sprintf("Are you sure you want to stop Volume %q? [y/N]", *job.ID)
		answer, err := c.Ui.Ask(question)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to parse answer: %v", err))
			return 1
		}

		if answer == "" || strings.ToLower(answer)[0] == 'n' {
			// No case
			c.Ui.Output("Cancelling Volume stop")
			return 0
		} else if strings.ToLower(answer)[0] == 'y' && len(answer) > 1 {
			// Non exact match yes
			c.Ui.Output("For confirmation, an exact ‘y’ is required.")
			return 0
		} else if answer != "y" {
			c.Ui.Output("No confirmation detected. For confirmation, an exact 'y' is required.")
			return 1
		}
	}

	// Invoke the stop
	evalID, _, err := client.Jobs().Deregister(*job.ID, nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error deregistering Volume: %s", err))
		return 1
	}

	// If we are stopping a periodic VSM there won't be an evalID.
	if evalID == "" {
		return 0
	}

	if detach {
		c.Ui.Output(evalID)
		return 0
	}

	// Start monitoring the stop eval
	mon := newMonitor(c.Ui, client, length)
	return mon.monitor(evalID, false)
}

// DeleteVsm to get delete Volume through a API call to m-apiserver
func DeleteVsm(vname string) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		fmt.Println("Error getting maya-api-server IP Address: %v", err)
		return err
	}
	url := addr + "/latest/volumes/delete/" + vname

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("http.NewRequest() error: : %v", err)
		return err
	}
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("http.Do() error: : %v", err)
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		fmt.Println("Status error: %v\n", http.StatusText(code))
		return err
	}
	fmt.Println("Initiated Volume-Delete request for volume:", string(vname))
	return nil
}

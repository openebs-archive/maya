package command

import (
	"fmt"
	"strings"
)

// VsmStopCommand is a command implementation struct
type VsmStopCommand struct {
	Meta
}

// Help shows helpText for a particular CLI command
func (c *VsmStopCommand) Help() string {
	helpText := `
Usage: maya volume stop [options] <volume>

  Stop an existing vsm. This command is used to signal allocations
  to shut down for the given vsm ID. Upon successful deregistraion,
  an interactive monitor session will start to display log lines as
  the vsm unwinds its allocations and completes shutting down.

Stop Options:

  -detach
    Return immediately instead of entering monitor mode. After the
    deregister command is submitted, a new evaluation ID is printed to the
    screen, which can be used to examine the evaluation using the eval-status
    command.
  
  -purge
    Purge is used to stop the vsm and purge it from the system. If not set, the
    vsm will still be queryable and will be purged by the garbage collector.
  
  -yes
    Automatic yes to prompts.

  -verbose
    Display full information.
`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmStopCommand) Synopsis() string {
	return "Stop a running Volume"
}

// Run holds the flag values for CLI subcommands
func (c *VsmStopCommand) Run(args []string) int {
	var detach, verbose, autoYes bool

	flags := c.Meta.FlagSet("volume stop", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.BoolVar(&detach, "detach", false, "")
	flags.BoolVar(&verbose, "verbose", false, "")
	flags.BoolVar(&autoYes, "yes", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
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

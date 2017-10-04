package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/nomad/structs"
)

const (
	// maxFailedTGs is the maximum number of task groups we show failure reasons
	// for before defering to eval-status
	maxFailedTGs = 5
)

// VsmListCommand is a command implementation struct
type VsmListCommand struct {
	Meta
	length    int
	evals     bool
	allAllocs bool
	verbose   bool
}

type ListStub struct {
	Items []struct {
		Metadata struct {
			Annotations struct {
				BeJivaVolumeOpenebsIoCount   string `json:"vsm.openebs.io/replica-count"`
				BeJivaVolumeOpenebsIoVolSize string `json:"vsm.openebs.io/volume-size"`
				Iqn                          string `json:"vsm.openebs.io/iqn"`
				Targetportal                 string `json:"vsm.openebs.io/targetportals"`
			} `json:"annotations"`
			CreationTimestamp interface{} `json:"creationTimestamp"`
			Name              string      `json:"name"`
		} `json:"metadata"`
		Spec struct {
			AccessModes interface{} `json:"AccessModes"`
			Capacity    interface{} `json:"Capacity"`
			ClaimRef    interface{} `json:"ClaimRef"`
			OpenEBS     struct {
				VolumeID string `json:"volumeID"`
			} `json:"OpenEBS"`
			PersistentVolumeReclaimPolicy string `json:"PersistentVolumeReclaimPolicy"`
			StorageClassName              string `json:"StorageClassName"`
		} `json:"spec"`
		Status struct {
			Message string `json:"Message"`
			Phase   string `json:"Phase"`
			Reason  string `json:"Reason"`
		} `json:"status"`
	} `json:"items"`
	Metadata struct {
	} `json:"metadata"`
}

// Help shows helpText for a particular CLI command
func (c *VsmListCommand) Help() string {
	helpText := `
Usage: maya volume list [options]

This command displays status of available Volumes. 
If no volume ID is given, a list of all known volume will be dumped.

Volume list options:

    -short
      Display short output. Used only when a single job is being
      queried, and drops verbose information about allocations.

    -evals
      Display the evaluations associated with the job.

    -all-allocs
      Display all allocations matching the job ID, including those
      from an older instance of the job.

    -verbose
      Display full information.

`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmListCommand) Synopsis() string {
	return "Display status information about Volume(s)"
}

// Run holds the flag values for CLI subcommands
func (c *VsmListCommand) Run(args []string) int {
	var short bool

	flags := c.Meta.FlagSet("volume list", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.BoolVar(&short, "short", false, "")
	flags.BoolVar(&c.evals, "evals", false, "")
	flags.BoolVar(&c.allAllocs, "all-allocs", false, "")
	flags.BoolVar(&c.verbose, "verbose", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we either got no jobs or exactly one.
	args = flags.Args()
	if len(args) > 1 {
		c.Ui.Error(c.Help())
		return 1
	}
	//TODO
	addr := os.Getenv("KUBERNETES_SERVICE_HOST")
	if addr != "" {
		VsmListOutput()
		return 0
	}

	// Truncate the id unless full length is requested
	c.length = shortId
	if c.verbose {
		c.length = fullId
	}

	// Get the HTTP client
	client, err := c.Meta.Client()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	// Invoke list mode if no job ID.
	if len(args) == 0 {
		jobs, _, err := client.Jobs().List(nil)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error querying jobs: %s", err))
			return 1
		}

		if len(jobs) == 0 {
			// No output if we have no volumes
			c.Ui.Output("No Volumes are running")
		} else {
			c.Ui.Output(createVsmListOutput(jobs))
		}
		return 0
	}

	// Try querying the job
	jobID := args[0]
	jobs, _, err := client.Jobs().PrefixList(jobID)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying job: %s", err))
		return 1
	}
	if len(jobs) == 0 {
		c.Ui.Error(fmt.Sprintf("No job(s) with prefix or id %q found", jobID))
		return 1
	}
	if len(jobs) > 1 && strings.TrimSpace(jobID) != jobs[0].ID {
		c.Ui.Output(fmt.Sprintf("Prefix matched multiple jobs\n\n%s", createVsmListOutput(jobs)))
		return 0
	}
	// Prefix lookup matched a single job
	job, _, err := client.Jobs().Info(jobs[0].ID, nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying job: %s", err))
		return 1
	}

	periodic := job.IsPeriodic()
	parameterized := job.IsParameterized()

	// Format the job info
	basic := []string{
		fmt.Sprintf("ID|%s", *job.ID),
		fmt.Sprintf("Name|%s", *job.Name),
		fmt.Sprintf("Type|%s", *job.Type),
		fmt.Sprintf("Priority|%d", *job.Priority),
		fmt.Sprintf("Datacenters|%s", strings.Join(job.Datacenters, ",")),
		fmt.Sprintf("Status|%s", *job.Status),
		fmt.Sprintf("Periodic|%v", periodic),
		fmt.Sprintf("Parameterized|%v", parameterized),
	}

	if periodic && !parameterized {
		location, err := job.Periodic.GetLocation()
		if err == nil {
			now := time.Now().In(location)
			next := job.Periodic.Next(now)
			basic = append(basic, fmt.Sprintf("Next Periodic Launch|%s",
				fmt.Sprintf("%s (%s from now)",
					formatTime(next), formatTimeDifference(now, next, time.Second))))
		}
	}

	c.Ui.Output(formatKV(basic))

	// Exit early
	if short {
		return 0
	}

	// Print periodic job information
	if periodic && !parameterized {
		if err := c.outputPeriodicInfo(client, job); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	} else if parameterized {
		if err := c.outputParameterizedInfo(client, job); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	} else {
		if err := c.outputJobInfo(client, job); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	}

	return 0
}

// outputPeriodicInfo prints information about the passed periodic job. If a
// request fails, an error is returned.
func (c *VsmListCommand) outputPeriodicInfo(client *api.Client, job *api.Job) error {
	// Output the summary
	if err := c.outputJobSummary(client, job); err != nil {
		return err
	}

	// Generate the prefix that matches launched jobs from the periodic job.
	prefix := fmt.Sprintf("%s%s", *job.ID, structs.PeriodicLaunchSuffix)
	children, _, err := client.Jobs().PrefixList(prefix)
	if err != nil {
		return fmt.Errorf("Error querying job: %s", err)
	}

	if len(children) == 0 {
		c.Ui.Output("\nNo instances of periodic job found")
		return nil
	}

	out := make([]string, 1)
	out[0] = "ID|Status"
	for _, child := range children {
		// Ensure that we are only showing jobs whose parent is the requested
		// job.
		if child.ParentID != *job.ID {
			continue
		}

		out = append(out, fmt.Sprintf("%s|%s",
			child.ID,
			child.Status))
	}

	c.Ui.Output(c.Colorize().Color("\n[bold]Previously Launched Jobs[reset]"))
	c.Ui.Output(formatList(out))
	return nil
}

// outputParameterizedInfo prints information about a parameterized job. If a
// request fails, an error is returned.
func (c *VsmListCommand) outputParameterizedInfo(client *api.Client, job *api.Job) error {
	// Output parameterized job details
	c.Ui.Output(c.Colorize().Color("\n[bold]Parameterized Job[reset]"))
	parameterizedJob := make([]string, 3)
	parameterizedJob[0] = fmt.Sprintf("Payload|%s", job.ParameterizedJob.Payload)
	parameterizedJob[1] = fmt.Sprintf("Required Metadata|%v", strings.Join(job.ParameterizedJob.MetaRequired, ", "))
	parameterizedJob[2] = fmt.Sprintf("Optional Metadata|%v", strings.Join(job.ParameterizedJob.MetaOptional, ", "))
	c.Ui.Output(formatKV(parameterizedJob))

	// Output the summary
	if err := c.outputJobSummary(client, job); err != nil {
		return err
	}

	// Generate the prefix that matches launched jobs from the parameterized job.
	prefix := fmt.Sprintf("%s%s", *job.ID, structs.DispatchLaunchSuffix)
	children, _, err := client.Jobs().PrefixList(prefix)
	if err != nil {
		return fmt.Errorf("Error querying job: %s", err)
	}

	if len(children) == 0 {
		c.Ui.Output("\nNo dispatched instances of parameterized job found")
		return nil
	}

	out := make([]string, 1)
	out[0] = "ID|Status"
	for _, child := range children {
		// Ensure that we are only showing jobs whose parent is the requested
		// job.
		if child.ParentID != *job.ID {
			continue
		}

		out = append(out, fmt.Sprintf("%s|%s",
			child.ID,
			child.Status))
	}

	c.Ui.Output(c.Colorize().Color("\n[bold]Dispatched Jobs[reset]"))
	c.Ui.Output(formatList(out))
	return nil
}

// outputJobInfo prints information about the passed non-periodic job. If a
// request fails, an error is returned.
func (c *VsmListCommand) outputJobInfo(client *api.Client, job *api.Job) error {
	var evals, allocs []string

	// Query the allocations
	jobAllocs, _, err := client.Jobs().Allocations(*job.ID, c.allAllocs, nil)
	if err != nil {
		return fmt.Errorf("Error querying job allocations: %s", err)
	}

	// Query the evaluations
	jobEvals, _, err := client.Jobs().Evaluations(*job.ID, nil)
	if err != nil {
		return fmt.Errorf("Error querying job evaluations: %s", err)
	}

	// Output the summary
	if err := c.outputJobSummary(client, job); err != nil {
		return err
	}

	// Determine latest evaluation with failures whose follow up hasn't
	// completed, this is done while formatting
	var latestFailedPlacement *api.Evaluation
	blockedEval := false

	// Format the evals
	evals = make([]string, len(jobEvals)+1)
	evals[0] = "ID|Priority|Triggered By|Status|Placement Failures"
	for i, eval := range jobEvals {
		failures, _ := evalFailureStatus(eval)
		evals[i+1] = fmt.Sprintf("%s|%d|%s|%s|%s",
			limit(eval.ID, c.length),
			eval.Priority,
			eval.TriggeredBy,
			eval.Status,
			failures,
		)

		if eval.Status == "blocked" {
			blockedEval = true
		}

		if len(eval.FailedTGAllocs) == 0 {
			// Skip evals without failures
			continue
		}

		if latestFailedPlacement == nil || latestFailedPlacement.CreateIndex < eval.CreateIndex {
			latestFailedPlacement = eval
		}
	}

	if c.verbose || c.evals {
		c.Ui.Output(c.Colorize().Color("\n[bold]Evaluations[reset]"))
		c.Ui.Output(formatList(evals))
	}

	if blockedEval && latestFailedPlacement != nil {
		c.outputFailedPlacements(latestFailedPlacement)
	}

	// Format the allocs
	c.Ui.Output(c.Colorize().Color("\n[bold]Allocations[reset]"))
	if len(jobAllocs) > 0 {
		allocs = make([]string, len(jobAllocs)+1)
		allocs[0] = "ID|Eval ID|Node ID|Task Group|Desired|Status|Created At"
		for i, alloc := range jobAllocs {
			allocs[i+1] = fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
				limit(alloc.ID, c.length),
				limit(alloc.EvalID, c.length),
				limit(alloc.NodeID, c.length),
				alloc.TaskGroup,
				alloc.DesiredStatus,
				alloc.ClientStatus,
				formatUnixNanoTime(alloc.CreateTime))
		}

		c.Ui.Output(formatList(allocs))
	} else {
		c.Ui.Output("No allocations placed")
	}
	return nil
}

// outputJobSummary displays the given jobs summary and children job summary
// where appropriate
func (c *VsmListCommand) outputJobSummary(client *api.Client, job *api.Job) error {
	// Query the summary
	summary, _, err := client.Jobs().Summary(*job.ID, nil)
	if err != nil {
		return fmt.Errorf("Error querying job summary: %s", err)
	}

	if summary == nil {
		return nil
	}

	periodic := job.IsPeriodic()
	parameterizedJob := job.IsParameterized()

	// Print the summary
	if !periodic && !parameterizedJob {
		c.Ui.Output(c.Colorize().Color("\n[bold]Summary[reset]"))
		summaries := make([]string, len(summary.Summary)+1)
		summaries[0] = "Task Group|Queued|Starting|Running|Failed|Complete|Lost"
		taskGroups := make([]string, 0, len(summary.Summary))
		for taskGroup := range summary.Summary {
			taskGroups = append(taskGroups, taskGroup)
		}
		sort.Strings(taskGroups)
		for idx, taskGroup := range taskGroups {
			tgs := summary.Summary[taskGroup]
			summaries[idx+1] = fmt.Sprintf("%s|%d|%d|%d|%d|%d|%d",
				taskGroup, tgs.Queued, tgs.Starting,
				tgs.Running, tgs.Failed,
				tgs.Complete, tgs.Lost,
			)
		}
		c.Ui.Output(formatList(summaries))
	}

	// Always display the summary if we are periodic or parameterized, but
	// only display if the summary is non-zero on normal jobs
	if summary.Children != nil && (parameterizedJob || periodic || summary.Children.Sum() > 0) {
		if parameterizedJob {
			c.Ui.Output(c.Colorize().Color("\n[bold]Parameterized Job Summary[reset]"))
		} else {
			c.Ui.Output(c.Colorize().Color("\n[bold]Children Job Summary[reset]"))
		}
		summaries := make([]string, 2)
		summaries[0] = "Pending|Running|Dead"
		summaries[1] = fmt.Sprintf("%d|%d|%d",
			summary.Children.Pending, summary.Children.Running, summary.Children.Dead)
		c.Ui.Output(formatList(summaries))
	}

	return nil
}

func (c *VsmListCommand) outputFailedPlacements(failedEval *api.Evaluation) {
	if failedEval == nil || len(failedEval.FailedTGAllocs) == 0 {
		return
	}

	c.Ui.Output(c.Colorize().Color("\n[bold]Placement Failure[reset]"))

	sorted := sortedTaskGroupFromMetrics(failedEval.FailedTGAllocs)
	for i, tg := range sorted {
		if i >= maxFailedTGs {
			break
		}

		c.Ui.Output(fmt.Sprintf("Task Group %q:", tg))
		metrics := failedEval.FailedTGAllocs[tg]
		c.Ui.Output(formatAllocMetrics(metrics, false, "  "))
		if i != len(sorted)-1 {
			c.Ui.Output("")
		}
	}

	if len(sorted) > maxFailedTGs {
		trunc := fmt.Sprintf("\nPlacement failures truncated. To see remainder run:\nnomad eval-status %s", failedEval.ID)
		c.Ui.Output(trunc)
	}
}

// list general information about a list of jobs
func createVsmListOutput(jobs []*api.JobListStub) string {
	out := make([]string, len(jobs)+1)
	out[0] = "ID|Type|Priority|Status"
	for i, job := range jobs {
		out[i+1] = fmt.Sprintf("%s|%s|%d|%s",
			job.ID,
			job.Type,
			job.Priority,
			job.Status)
	}
	return formatList(out)
}

func GetVsm(obj interface{}) error {

	body, err := RestClient()
	if err != nil {
		fmt.Sprintf("Error querying Volumes: %s", err)
		return err
	}
	return Parser(body, obj)

}

func VsmListOutput() error {

	var vsms ListStub
	GetVsm(&vsms)
	out := make([]string, len(vsms.Items)+1)
	out[0] = "Name|Status"
	for i, items := range vsms.Items {
		if items.Status.Reason == "" {
			items.Status.Reason = "Running"
		}
		out[i+1] = fmt.Sprintf("%s|%s",
			items.Metadata.Name,
			items.Status.Reason)
	}
	if len(out) == 1 {
		fmt.Println("No Volumes are running")
		return nil
	}
	fmt.Println(formatList(out))
	return nil

}

func RestClient() ([]byte, error) {
	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		fmt.Println(err)
		return nil, err
	}
	url := addr + "/latest/volumes/"
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			fmt.Println("Volume not found at M_API server")
			return nil, err
		} else if resp.StatusCode == 503 {
			fmt.Println("M_API server not reachable")
			return nil, err
		}
	} else {
		fmt.Println("M_API server not reachable")
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return []byte(body), err

	//	return json.NewDecoder(resp.Body).Decode(obj), nil
	//	return resp.Body, nil
}

func Parser(body []byte, obj interface{}) error {
	err := json.Unmarshal(body, &obj)
	if err != nil {
		fmt.Println("Error", err)
	}
	return err
}

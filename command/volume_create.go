package command

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/openebs/maya/types/v1"
	yaml "gopkg.in/yaml.v2"
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
		if strings.HasSuffix(c.size, "G") != true {
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

		resp := CreateAPIVsm(c.vsmname, c.size)
		if resp != nil {
			c.Ui.Error(fmt.Sprintf("Error Creating Volume %v", resp))
		}
	}
	return op
}

// CreateAPIVsm to create the Vsm through a API call to m-apiserver
func CreateAPIVsm(vname string, size string) error {

	var vs v1.VolumeAPISpec

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		fmt.Println(err)
		return err
	}
	url := addr + "/latest/volumes/"

	vs.Kind = "PersistentVolumeClaim"
	vs.APIVersion = "v1"
	vs.Metadata.Name = vname
	vs.Metadata.Labels.Storage = size

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(vs)

	fmt.Printf("Volume Spec Created:\n%v\n", string(yamlValue))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return err
	}

	req.Header.Add("Content-Type", "application/yaml")

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return err
	}
	code := resp.StatusCode

	if code != http.StatusOK {

		fmt.Printf("Status error: %v\n", http.StatusText(code))
		os.Exit(1)
	}

	fmt.Printf("Volume Successfully Created:\n%v\n", string(data))

	return nil
}

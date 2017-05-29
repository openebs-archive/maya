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

	yaml "gopkg.in/yaml.v2"
)

// VsmCreateCommand is a command implementation struct
type VsmCreateCommand struct {
	// To control this CLI's display
	M Meta
	// OS command to execute; <optional>
	Cmd     *exec.Cmd
	vsmname string
	size    string
}

// VsmSpec holds the config for creating a VSM
type VsmSpec struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		AccessModes []string `yaml:"accessModes"`
		Resources   struct {
			Requests struct {
				Storage string `yaml:"storage"`
			} `yaml:"requests"`
		} `yaml:"resources"`
	} `yaml:"spec"`
}

// Help shows helpText for a particular CLI command
func (c *VsmCreateCommand) Help() string {
	helpText := `
Usage: maya vsm-create [options] <path>

  Creates a new VSM using the specification located at <path>.

  On successful vsm creation submission and scheduling, exit code 0 will be
  returned. If there are placement issues encountered
  (unsatisfiable constraints, resource exhaustion, etc), then the
  exit code will be 2. Any other errors, including client connection
  issues or internal errors, are indicated by exit code 1.

VSM Create Options:
  -name
    Name of the vsm
  -size
    Provisioning size of the vsm(default is 5G)
`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmCreateCommand) Synopsis() string {
	return "Creates a new VSM"
}

// Run to get the flag values and start execution
//The logic of this function can be understood by understanding
// the help text defined earlier.
func (c *VsmCreateCommand) Run(args []string) int {

	var op int

	flags := c.M.FlagSet("vsm-create", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }
	flags.StringVar(&c.vsmname, "name", "", "")
	flags.StringVar(&c.size, "size", "5G", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// specs file is mandatory
	args = flags.Args()
	if len(args) != 1 && len(strings.TrimSpace(c.vsmname)) == 0 {
		c.M.Ui.Error(c.Help())
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
			Ui:  c.M.Ui,
		}

		if op = ic.Execute(); 0 != op {
			c.M.Ui.Error("Error creating vsm")
			return op
		}
		return 1
	}
	if c.vsmname != " " {
		if strings.HasSuffix(c.size, "G") != true {
			fmt.Println("-size should contain the suffix 'G',which represent the size in GB (exp: 10G)")
			return 0
		}
		err := CreateAPIVsm(c.vsmname, c.size)
		if err != nil {
			fmt.Println("Error Creating Vsm")
		}
	}
	return op
}

// CreateAPIVsm to create the Vsm through a API call to m-apiserver
func CreateAPIVsm(vname string, size string) error {

	var vs VsmSpec

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		fmt.Println(err)
		return err
	}
	url := addr + "/latest/volumes/"

	vs.Metadata.Name = vname
	vs.Spec.Resources.Requests.Storage = size

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(vs)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))

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

	fmt.Printf("VSM Successfully Created:\n%v\n", string(data))

	return err
}

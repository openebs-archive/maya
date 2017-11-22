package command

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"

	//"github.com/rancher/go-rancher/client"
	"github.com/openebs/maya/pkg/client/jiva"
)

var (
	MaximumVolumeNameSize = 64
	parsePattern          = regexp.MustCompile(`(.*):(\d+)`)
)

// SnapshotCreateCommand is a command implementation struct
type SnapshotCreateCommand struct {
	Meta
	Name   string
	Sname  string
	Labels map[string]string
}

// StringSlice is an opaque type for []string to satisfy flag.Value
type StringSlice []string

// Set appends the string value to the list of values
func (f *StringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *StringSlice) String() string {
	return fmt.Sprintf("%s", *f)
}

// Value returns the slice of strings set by this flag
func (f *StringSlice) Value() []string {
	return *f
}

// Help shows helpText for a particular CLI command
func (c *SnapshotCreateCommand) Help() string {
	helpText := `
	Usage: maya snapshot create -volname <vol> -snapname <snap>

	This command will create a snapshot of a given Volume.

	`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *SnapshotCreateCommand) Synopsis() string {
	return "Creates snapshot of a Volume"
}

// Run holds the flag values for CLI subcommands
func (c *SnapshotCreateCommand) Run(args []string) int {
	var (
		labelMap map[string]string
		err      error
	)

	flags := c.Meta.FlagSet("snapshot", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	flags.StringVar(&c.Name, "volname", "", "")
	flags.StringVar(&c.Sname, "snapname", "", "")
	//flags.String(&c.Labels, "label", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}
	/* var name string
	if len(c.Args()) > 0 {
		name = c.Args()[0]
	} */
	/*	var flagset *flag.FlagSet
		labels := lookupStringSlice("label", flagset)
		fmt.Sprint(labels)
		if labels != nil {
			labelMap, err = ParseLabels(labels)
			if err != nil {
				fmt.Printf("cannot parse backup labels")
				return 1
			}
		}
	*/
	//	str := os.Args[1:]
	//	labelMap = map[str]string
	//var client ControllerClient

	fmt.Println("Creating Snapshot of Volume :", c.Name)
	id, err := Snapshot(c.Name, c.Sname, labelMap)
	if err != nil {
		log.Fatalf("Error running create snapshot command: %v", err)
		return 1
	}

	fmt.Println("Created Snapshot is:", id)
	return 0

}

// Snapshot is used to get a snapshot
func Snapshot(volname string, snapname string, labels map[string]string) (string, error) {

	annotations, err := GetVolAnnotations(volname)
	if err != nil || annotations == nil {

		return "", err
	}

	if annotations.ControllerStatus != "Running" {
		fmt.Println("Volume not reachable")
		return "", err
	}
	controller, err := client.NewControllerClient(annotations.ControllerIP + ":9501")

	if err != nil {
		return "", err
	}

	volume, err := client.GetVolume(controller.Address)
	if err != nil {
		return "", err
	}

	url := controller.Address + "/volumes/" + volume.Id + "?action=snapshot"

	input := client.SnapshotInput{
		Name:   snapname,
		Labels: labels,
	}
	output := client.SnapshotOutput{}
	err = controller.Post(url, input, &output)
	if err != nil {
		return "", err
	}

	return output.Id, err
}

// ParseLabels helper to parse array string and return a
// map[string]string key:value pair
func ParseLabels(labels []string) (map[string]string, error) {
	result := map[string]string{}
	for _, label := range labels {
		kv := strings.Split(label, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Invalid label not in <key>=<value> format %v", label)
		}
		key := kv[0]
		value := kv[1]
		//Well, we should rename that ValidVolumeName
		if !ValidVolumeName(key) {
			return nil, fmt.Errorf("Invalid key %v for label %v", key, label)
		}
		if !ValidVolumeName(value) {
			return nil, fmt.Errorf("Invalid value %v for label %v", value, label)
		}
		result[key] = value
	}
	return result, nil
}

// ValidVolumeName is used to validate volume name
func ValidVolumeName(name string) bool {
	if len(name) > MaximumVolumeNameSize {
		return false
	}
	validName := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)
	return validName.MatchString(name)
}

// lookupStringSlice is used to look up string slice
func lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*StringSlice)).Value()

	}

	return nil
}

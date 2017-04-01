package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type VsmStatsCommand struct {
	Meta
	address string
	host    string
	length  int
}

func (c *VsmStatsCommand) Help() string {
	helpText := `
Usage: maya vsm-stats [options] <vsm>

  Display stats information about VSM.

  Stats Options:
`
	return strings.TrimSpace(helpText)
}

func (c *VsmStatsCommand) Synopsis() string {
	return "Display stats information about Vsm(s)"
}

func (c *VsmStatsCommand) Run(args []string) int {

	flags := c.Meta.FlagSet("vsm-stats", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	// Check that we either got no or exactly one replica.
	args = flags.Args()
	if len(args) > 1 {
		c.Ui.Error(c.Help())
		return 1
	}
	var path string
	path = "/tmp/demo-vsm1-vol1/be1/revision.counter"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			path = "/tmp/demo-vsm1-vol1/be0/revision.counter"
		} else {

		}
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	// ioutil.ReadAll() will read every byte
	// from the reader (in this case a file),
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	str := string(data) // convert content to a 'string'
	fmt.Println("Revision counter:")
	fmt.Println(str) // print the content as a 'string'

	return 0
}

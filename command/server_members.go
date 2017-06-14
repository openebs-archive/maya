package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/ryanuber/columnize"
)

// ServerMembersCommand is basic struct for cli command
type ServerMembersCommand struct {
	Meta
	Cmd *exec.Cmd
}

// Help is helper function
func (c *ServerMembersCommand) Help() string {
	helpText := `
Usage: maya omm-status [options]

  Display a list of the known servers and their status. Only Nomad servers are
  able to service this command.

General Options:

  ` + generalOptionsUsage() + `

Server Members Options:

  -detailed
    Show detailed information about each member. This dumps
    a raw set of tags which shows more information than the
    default output format.
`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *ServerMembersCommand) Synopsis() string {
	return "Display a list of known servers and their status"
}

// Run holds the flag values for CLI subcommands
func (c *ServerMembersCommand) Run(args []string) int {
	var detailed bool

	flags := c.Meta.FlagSet("server-members", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.BoolVar(&detailed, "detailed", false, "Show detailed output")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check for extra arguments
	args = flags.Args()
	if len(args) != 0 {
		c.Ui.Error(c.Help())
		return 1
	}

	// Get the HTTP client
	client, err := c.Meta.Client()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	// Query the members
	srvMembers, err := client.Agent().Members()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying servers: %s", err))
		return 1
	}

	if srvMembers == nil {
		c.Ui.Error("Agent doesn't know about server members")
		return 0
	}

	// Sort the members
	sort.Sort(api.AgentMembersNameSort(srvMembers.Members))

	// Determine the leaders per region.
	leaders, err := regionLeaders(client, srvMembers.Members)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error determining leaders: %s", err))
		return 1
	}

	// Format the list
	var out []string
	if detailed {
		out = detailedOutput(srvMembers.Members)
	} else {
		out = standardOutput(srvMembers.Members, leaders)
	}

	// Dump the list
	c.Ui.Output(columnize.SimpleFormat(out))

	//getting the m-apiserver env variable
	addr := os.Getenv("MAPI_ADDR")

	_, resp := c.mserverStatus()

	if resp != nil {
		fmt.Printf("\nM-apiserver Status : %v\n", resp)
		return 0
	}

	fmt.Printf("\nM-apiserver listening at : %v\n", addr)

	return 0
}

func standardOutput(mem []*api.AgentMember, leaders map[string]string) []string {
	// Format the members list
	members := make([]string, len(mem)+1)
	members[0] = "Name|Address|Port|Status|Leader|Protocol|Build|Datacenter|Region"
	for i, member := range mem {
		reg := member.Tags["region"]
		regLeader, ok := leaders[reg]
		isLeader := false
		if ok {
			if regLeader == net.JoinHostPort(member.Addr, member.Tags["port"]) {

				isLeader = true
			}
		}

		members[i+1] = fmt.Sprintf("%s|%s|%d|%s|%t|%d|%s|%s|%s",
			member.Name,
			member.Addr,
			member.Port,
			member.Status,
			isLeader,
			member.ProtocolCur,
			member.Tags["build"],
			member.Tags["dc"],
			member.Tags["region"])
	}
	return members
}

func detailedOutput(mem []*api.AgentMember) []string {
	// Format the members list
	members := make([]string, len(mem)+1)
	members[0] = "Name|Address|Port|Tags"
	for i, member := range mem {
		// Format the tags
		tagPairs := make([]string, 0, len(member.Tags))
		for k, v := range member.Tags {
			tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", k, v))
		}
		tags := strings.Join(tagPairs, ",")

		members[i+1] = fmt.Sprintf("%s|%s|%d|%s",
			member.Name,
			member.Addr,
			member.Port,
			tags)
	}
	return members
}

// regionLeaders returns a map of regions to the IP of the member that is the
// leader.
func regionLeaders(client *api.Client, mem []*api.AgentMember) (map[string]string, error) {
	// Determine the unique regions.
	leaders := make(map[string]string)
	regions := make(map[string]struct{})
	for _, m := range mem {
		regions[m.Tags["region"]] = struct{}{}
	}

	if len(regions) == 0 {
		return leaders, nil
	}

	status := client.Status()
	for reg := range regions {
		l, err := status.RegionLeader(reg)
		if err != nil {
			// This error means that region has no leader.
			if strings.Contains(err.Error(), "No cluster leader") {
				continue
			}
			return nil, err
		}

		leaders[reg] = l
	}

	return leaders, nil
}

// mserverStatus to get the status of mayaserver deamon,
// TODO proper CLI command once mayaserver have it's own
func (c *ServerMembersCommand) mserverStatus() (string, error) {

	//getting the m-apiserver env variable
	addr := os.Getenv("MAPI_ADDR")

	var url bytes.Buffer
	url.WriteString(addr + "/latest/meta-data/instance-id")
	resp, err := http.Get(url.String())

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(body[:]), err

}

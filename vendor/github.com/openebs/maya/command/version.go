package command

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mitchellh/cli"
)

// VersionCommand is a Command implementation prints the version.
type VersionCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

// Help shows helpText for a particular CLI command
func (c *VersionCommand) Help() string {
	helpText := `
Usage: maya version
							           
This command provides versioning and other details relevant to maya.

`
	return strings.TrimSpace(helpText)
}

// Run holds the flag values for CLI subcommands
func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer
	var s *ServerMembersCommand
	fmt.Fprintf(&versionString, "Maya v%s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", c.VersionPrerelease)

		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}

	c.Ui.Output(versionString.String())

	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("OS/Arch:", runtime.GOOS, "/", runtime.GOARCH)
	addr := os.Getenv("MAPI_ADDR")

	/*if addr == "" {
		addr = getEnvOrDefault(addr)

		os.Setenv("MAPI_ADDR", addr)
		addr = os.Getenv("MAPI_ADDR")

	}*/
	_, err := s.mserverStatus()
	if err != nil {
		fmt.Println("M-apiserver: Unable to contact M-apiserver :", addr)
	}
	if err == nil {
		fmt.Printf("M-apiserver: %v\n", addr)
	}
	_, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if !ok {
		for _, e := range os.Environ() {
			ok := strings.Contains(e, "NOMAD_ADDR")
			if !ok {
				fmt.Println("Provider : Unknown")
				return 0
			}
			fmt.Println("Provider : NOMAD")
			return 0
		}

	}
	fmt.Println("Provider: KUBERNETES")

	return 0
}

// Synopsis shows short information related to CLI command
func (c *VersionCommand) Synopsis() string {
	return "Prints version and other details relevant to maya"
}

/*
Copyright 2017 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of Ark, set by the go linker's -X flag at build time.
	Version string

	// GitSHA is the actual commit that is being built, set by the go linker's -X flag at build time.
	GitSHA string

	// GitTreeState indicates if the git tree is clean or dirty, set by the go linker's -X flag at build
	// time.
	GitTreeState string
)

// NewCommand creates the version command
func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print Maya version information",
		Long: `Print Maya version information for the current context

Example:
maya version
	`,

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", getVersion())
			fmt.Printf("Git commit: %s", getGitCommit())
			//fmt.Printf("Git tree state: %s\n", GitTreeState)
			fmt.Printf("Go-Version: %s\n", runtime.Version())
			fmt.Printf("GOARCH: %s\n", runtime.GOARCH)
			fmt.Printf("GOOS: %s\n", runtime.GOOS)

		},
	}

	return cmd
}

// FormattedGitSHA renders the Git SHA with an indicator of the tree state.
func FormattedGitSHA() string {
	if GitTreeState != "clean" {
		return fmt.Sprintf("%s-%s", GitSHA, GitTreeState)
	}
	return GitSHA
}

var (
	versionFile = "/src/github.com/openebs/maya/VERSION"
)

func getVersion() string {
	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return string(vBytes)
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// ignore error
		return ""
	}
	return string(output)
}

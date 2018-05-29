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
	"runtime"
	"strconv"
	"strings"

	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/version"
	"github.com/spf13/cobra"
)

// NewCommand creates the version command
func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints version and other details relevant to maya",
		Long: `Prints version and other details relevant to maya

Usage:
maya version
	`,

		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(checkLatestVersion(version.GetVersion()), util.Fatal)
			fmt.Printf("Version: %s\n",
				version.GetVersion()+version.GetBuildMeta())
			fmt.Printf("Git commit: %s\n", version.GetGitCommit())

			fmt.Printf("GO Version: %s\n", runtime.Version())
			fmt.Printf("GO ARCH: %s\n", runtime.GOARCH)
			fmt.Printf("GO OS: %s\n", runtime.GOOS)

			fmt.Println("m-apiserver url: ", mapiserver.GetURL())
			fmt.Println("m-apiserver status: ", mapiserver.GetConnectionStatus())

			fmt.Println("Provider: ", orchprovider.DetectOrchProviderFromEnv())

		},
	}

	return cmd
}

func checkLatestVersion(installedVersion string) error {
	if installedVersion == "" {
		return fmt.Errorf("GetVersion() returning empty string")
	}

	// removes first character i.e 'v' from the version
	installedVersion = installedVersion[1:]

	latestVersion, err := version.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("found error - %s", err)
	}

	latest := parseVersion(latestVersion)
	installed := parseVersion(installedVersion)

	if latest == nil || installed == nil {
		return fmt.Errorf("error in parsing string")
	}

	flag := false

	if installed[0] < latest[0] {
		flag = true
	} else if installed[1] < latest[1] && installed[0] == latest[0] {
		flag = true
	} else if installed[2] < latest[2] && installed[1] == latest[1] && installed[0] == latest[0] {
		flag = true
	}

	if flag == true {
		fmt.Println("A newer version of mayactl is available!")
		fmt.Printf("Installed Version: v%s\n", installedVersion)
		fmt.Printf("Latest version: v%s\n", latestVersion)
		fmt.Println()
	}

	return nil
}

func parseVersion(version string) []int64 {
	versionList := strings.Split(version, ".")
	versionNumber := []int64{}

	for _, v := range versionList {
		j, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Printf("Error - %s\n", err)
			return nil
		}
		versionNumber = append(versionNumber, j)
	}
	return versionNumber
}

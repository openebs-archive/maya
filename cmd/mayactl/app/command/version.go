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

	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/pkg/client/mapiserver"
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

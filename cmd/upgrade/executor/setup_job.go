/*
Copyright 2019 The OpenEBS Authors.

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

package executor

import (
	"flag"
	"fmt"
	//"os"
	"strings"

	//"github.com/golang/glog"
	"github.com/spf13/cobra"
	//errors "github.com/openebs/maya/pkg/errors/v1alpha1"
)

var (
	cmdName = "upgrade"
	usage   = fmt.Sprintf("%s", cmdName)
)

// NewJob will setup a new upgrade job
func NewJob() *cobra.Command {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "OpenEBS Upgrade Utility",
		Long: `An utility to uggrade OpenEBS Storage Pools and Volumes,
			run as a Kubernetes Job`,
		PersistentPreRun: PreRun,
	}

	cmd.AddCommand(
		NewUpgradeJivaVolumeJob(),
	)

	cmd.PersistentFlags().StringVarP(&options.fromVersion,
		"from-version", "",
		options.fromVersion,
		"current version of the resource (pool or volume) being upgraded.")

	cmd.PersistentFlags().StringVarP(&options.toVersion,
		"to-version", "",
		options.toVersion,
		"new version to which resource (pool or volume) should be upgraded.")

	cmd.PersistentFlags().StringVarP(&options.namespace,
		"openebs-namespace", "",
		options.namespace,
		"namespace where openebs is installed.")

	cmd.PersistentFlags().StringVarP(&options.imageUrlPrefix,
		"to-version-image-prefix", "",
		options.imageUrlPrefix,
		"custom image prefix, when not using the default.")

	cmd.PersistentFlags().StringVarP(&options.toVersionImageTag,
		"to-version-image-tag", "",
		options.toVersionImageTag,
		"custom image tag, when not using the default.")

	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Hack: Without the following line, the logs will be prefixed with Error
	_ = flag.CommandLine.Parse([]string{})

	return cmd
}

// PreRun will check for environement variables to be read and intialized.
func PreRun(cmd *cobra.Command, args []string) {
	namespace := getOpenEBSNamespace()
	if len(strings.TrimSpace(namespace)) == 0 {
		options.namespace = namespace
	}
}

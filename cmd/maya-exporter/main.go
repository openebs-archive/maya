// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/maya-exporter/app/command"
	mayalogger "github.com/openebs/maya/pkg/logs"
	"github.com/openebs/maya/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(version.NewVersionCollector("maya_exporter"))
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run maya-exporter
func run() error {
	// Init logging
	mayalogger.InitLogs()
	defer mayalogger.FlushLogs()

	// Create & execute new command
	cmd, err := command.NewCmdVolumeExporter()
	if err != nil {
		glog.Error("Can't execute the command, error :", err)
		return err
	}

	return cmd.Execute()
}

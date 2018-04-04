/*
Copyright 2018 The OpenEBS Authors.

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

package main

import (
	"os"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/app/command"
	cstorlogger "github.com/openebs/maya/pkg/logs"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run cstor-pool-mgmt
func run() error {
	// Init logging
	cstorlogger.InitLogs()
	defer cstorlogger.FlushLogs()

	// Create & execute new command
	cmd, err := command.NewCStorPoolMgmt()
	if err != nil {
		return err
	}

	return cmd.Execute()
}

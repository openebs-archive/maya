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

	"net/http"
	_ "net/http/pprof"


	"github.com/openebs/maya/cmd/cstor-volume-mgmt/app/command"
	//"github.com/openebs/maya/pkg/debug"
	cstorlogger "github.com/openebs/maya/pkg/logs"
	//"github.com/pkg/profile"
)

func main() {
	//Enable CPU profiling.
	//if debug.EnableCPUProfiling() {
	//	defer profile.Start(profile.ProfilePath(debug.GetProfilePath())).Stop()
	//}

	go func() {
		http.HandleFunc("/", serveHTTP)
		http.ListenAndServe(":9664", nil)
	}()

	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run cstor-volume-mgmt
func run() error {
	// Init logging
	cstorlogger.InitLogs()
	defer cstorlogger.FlushLogs()

	// Create & execute new command
	cmd, err := command.NewCStorVolumeMgmt()
	if err != nil {
		return err
	}

	return cmd.Execute()
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, world"))
}

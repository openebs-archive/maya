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
	"flag"
	"os"

	log "github.com/golang/glog"
	upgrade "github.com/openebs/maya/cmd/upgrade/app/v1alpha1"
)

func main() {
	flag.Set("logtostderr", "true")
	configPath := flag.String("config-path", "/etc/config/upgrade", "path to config file.")
	defer log.Flush()
	flag.Parse()

	runOptions := &upgrade.Upgrade{
		ConfigPath: *configPath,
	}
	err := runOptions.Run()

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}

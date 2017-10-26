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

package version

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// GitCommit that was compiled. This will be filled in by the compiler.
	GitCommit string

	// Version show the version number,fill in by the compiler
	Version string

	// VersionPrerelease is a pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionMeta string

	versionFile   = "/src/github.com/openebs/maya/VERSION"
	buildMetaFile = "/src/github.com/openebs/maya/BUILDMETA"
)

func GetVersion() string {
	if Version != "" {
		return Version
	}
	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return strings.TrimSpace(string(vBytes))
}

func GetBuildMeta() string {
	if VersionMeta != "" {
		return "-" + VersionMeta
	}
	path := filepath.Join(os.Getenv("GOPATH") + buildMetaFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return "-" + strings.TrimSpace(string(vBytes))
}

func GetGitCommit() string {
	if GitCommit != "" {
		return GitCommit
	}
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// ignore error
		return ""
	}
	return strings.TrimSpace(string(output))
}

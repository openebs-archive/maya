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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/openebs/maya/types/v1"
)

var (
	// GitCommit that was compiled. This will be filled in by the compiler.
	GitCommit string

	// Version show the version number,fill in by the compiler
	Version string

	// LatestVersion shows the latest version available
	LatestVersion string

	// VersionPrerelease is a pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionMeta string

	versionFile   = "/src/github.com/openebs/maya/VERSION"
	buildMetaFile = "/src/github.com/openebs/maya/BUILDMETA"

	// GitAPIAddr stores the address of the github api
	GitAPIAddr = "https://api.github.com"
)

const (
	httpTimeout = 5 * time.Second
)

type release struct {
	TagName string `json:"tag_name"`
}

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

// GetLatestRelease returns the latest version available
func GetLatestRelease() (string, error) {

	url := GitAPIAddr + v1.GitAPI

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Unable to make a GET request. \n Error - %s\n", err)
		return "", err
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request timeout occured. \n Error - %s\n", err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Unable to read body. \n Error - %s\n", err)
		return "", err
	}

	Release := []release{}

	err = json.Unmarshal(body, &Release)
	if err != nil {
		fmt.Printf("Unable to unmarshal json body of latest release. \n Error%s\n", err)
		return "", err
	}

	// To get the latest stabled version
	for _, r := range Release {
		if !strings.Contains(r.TagName, "RC") {
			LatestVersion = r.TagName
			break
		}
	}

	return LatestVersion, nil
}

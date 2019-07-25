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
	"fmt"
	"os"
	"strings"

	upgrade090to100 "github.com/openebs/maya/pkg/upgrade/0.9.0-1.0.0/v1alpha1"
	upgrade100to110 "github.com/openebs/maya/pkg/upgrade/1.0.0-1.1.0/v1alpha1"
)

func main() {
	from := os.Args[1]
	to := os.Args[2]
	kind := os.Args[3]
	name := os.Args[4]
	openebsNamespace := os.Args[5]
	urlPrefix := ""
	imageTag := ""
	if len(os.Args) == 7 {
		urlPrefix = os.Args[6]
	}
	if len(os.Args) == 8 {
		urlPrefix = os.Args[6]
		imageTag = os.Args[7]
	}

	fromVersion := strings.Split(from, "-")[0]
	toVersion := strings.Split(to, "-")[0]

	switch fromVersion + "-" + toVersion {
	case "0.9.0-1.0.0":
		err := upgrade090to100.Exec(kind, name, openebsNamespace)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "1.0.0-1.1.0":
		err := upgrade100to110.Exec(from, to, kind, name, openebsNamespace, urlPrefix, imageTag)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid from version %s or to version %s", from, to)
		os.Exit(1)
	}
	os.Exit(0)
}

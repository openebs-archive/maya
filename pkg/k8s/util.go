/*
Copyright 2018 The OpenEBS Authors

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
package k8s

import (
	"strings"
)

// converts "svc:mysvc, team:alpha" to map[string]string{"svc":"mysvc", "team":"alpha"}
func parseLables(cutomLables string) map[string]string {
	parsedLables := make(map[string]string, 0)
	splitedLables := strings.Split(cutomLables, ",")
	for _, splitedLable := range splitedLables {
		splitedLable = strings.TrimSpace(splitedLable)
		splitedValues := strings.Split(splitedLable, ",")
		parsedLables[splitedValues[0]] = splitedValues[1]
	}
	return parsedLables
}

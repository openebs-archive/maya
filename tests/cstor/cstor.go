/*
:q
Copyright 2019 The OpenEBS Authors
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

package cstor

import "flag"

var (
	// KubeConfigPath is the path to
	// the kubeconfig provided at runtime
	KubeConfigPath string
	// ReplicaCount is the value of
	// replica count provided at runtime
	ReplicaCount int
	// PoolCount is the value of
	// max pool count in spc
	PoolCount int
)

// ParseFlags gets the flag values at run time
func ParseFlags() {
	flag.StringVar(&KubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
	flag.IntVar(&ReplicaCount, "cstor-replicas", 1, "value of replica count")
	flag.IntVar(&PoolCount, "cstor-maxpools", 1, "value of maxpool count")
}

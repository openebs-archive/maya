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

package environments

// Environment is the interface which integrate all the functionalities
// that environments like minikube, docker etc should have.
type Environment interface {
	Name() string
	Setup() error // If Setup of environment fails it is better to exit from there
	Status() (map[string]string, error)
	Teardown() error
}

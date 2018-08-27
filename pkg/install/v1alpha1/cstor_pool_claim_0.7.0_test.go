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
package v1alpha1

import (
	"os"
	"testing"
)

func setEnv() {
	os.Setenv(string(CASDefaultCstorPool), "true")
}

func unsetEnv() {
	os.Unsetenv(string(CASDefaultCstorPool))
}

func TestIsCstorSparsePoolEnabled(t *testing.T) {
	// Set env variable to enable sparse pool creation
	setEnv()
	result := IsCstorSparsePoolEnabled()
	if result != true {
		t.Errorf("Test failed as the env variable for cstor sparse pool creation is true but function returned false")
	}
	// Unset env variable to disable sparse pool creation
	unsetEnv()
	result = IsCstorSparsePoolEnabled()
	if result == true {
		t.Errorf("Test failed as the env variable for cstor sparse pool creation is unset but function returned true")
	}
}

func TestCstorPoolSpc070(t *testing.T) {
	// Set Env variable to enable sparse pool creation
	setEnv()
	listItems := CstorPoolSpc070()
	if len(listItems.Items) == 0 {
		t.Errorf("Expected non empty string list")
	}
	// Unset env variable to disable sparse pool creation
	unsetEnv()
	listItems = CstorPoolSpc070()
	if len(listItems.Items) != 0 {
		t.Errorf("Expected empty string list")
	}
}

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

package patch

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// ClientSet struct holds kubernetes and openebs clientsets.
type ClientSet struct {
	// kubeclientset is a standard kubernetes clientset
	Kubeclientset kubernetes.Interface
	// clientset is a openebs custom resource package generated for custom API group.
	OpenebsClientset clientset.Interface
}

// Patch is the struct based on standards of JSON patch.
type Patch struct {
	// Op defines the operation
	Op string `json:"op"`
	// Path defines the key path
	// eg. for
	// {
	//  	"Name": "openebs"
	//	    Category: {
	//		  "Inclusive": "v1",
	//		  "Rank": "A"
	//	     }
	// }
	// The path of 'Inclusive' would be
	// "/Name/Category/Inclusive"
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// Patcher interface has Patch function which can be implemented for several objects that needs to be patched.
type Patcher interface {
	Patch(string, types.PatchType) (interface{}, error)
}

// NewPatchPayload constructs the patch payload fo any type of object.
func NewPatchPayload(operation string, path string, value interface{}) (payload []Patch) {
	PatchPayload := make([]Patch, 1)
	PatchPayload[0].Op = operation
	PatchPayload[0].Path = path
	PatchPayload[0].Value = value
	return PatchPayload
}

// PatchCsp will patch the CSP object
func (c *ClientSet) PatchCsp(cspName string, patchType types.PatchType, patches []byte) (*apis.CStorPool, error) {
	csp, err := c.OpenebsClientset.OpenebsV1alpha1().CStorPools().Patch(cspName, patchType, patches)
	return csp, err
}

/*
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

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/catalog/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ToUnstructResources returns a list of unstructured instances
// corresponding to provided catalog's resources
func ToUnstructResources(c *apis.Catalog) ([]*unstructured.Unstructured, error) {
	var unstructs []*unstructured.Unstructured
	if c == nil {
		return nil, errors.New("failed to build unstruct instances for catalog resources: nil catalog")
	}
	for _, resource := range c.Spec.Items {
		u, err := unstruct.Unmarshal(resource.Template)
		if err != nil {
			return nil, err
		}
		unstructs = append(unstructs, u)
	}
	return unstructs, nil
}

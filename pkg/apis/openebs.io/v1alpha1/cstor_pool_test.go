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
	"strings"
	"testing"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCstorPoolString(t *testing.T) {
	tests := map[string]struct {
		csp                 *CStorPool
		expectedStringParts []string
	}{
		"cstor pool mycsp": {
			&CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "mycsp", Namespace: "default"},
				Spec: CStorPoolSpec{
					PoolSpec: CStorPoolAttr{
						CacheFile: "/tmp/.cache",
					},
				},
			},
			[]string{"cacheFile: /tmp/.cache", "name: mycsp", "namespace: default"},
		},
		"cstor pool mygoodcsp": {
			&CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "mygoodcsp", Namespace: "openebs"},
				Spec: CStorPoolSpec{
					PoolSpec: CStorPoolAttr{
						CacheFile: "/tmp/.cache2",
					},
				},
			},
			[]string{"cacheFile: /tmp/.cache2", "name: mygoodcsp", "namespace: openebs"},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.csp.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestCstorPoolJSONIndent(t *testing.T) {
	tests := map[string]struct {
		context             string
		csp                 *CStorPool
		expectedStringParts []string
	}{
		"my cstor pool": {
			"my cstor pool",
			&CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "mycsp", Namespace: "default"},
				Spec: CStorPoolSpec{
					PoolSpec: CStorPoolAttr{
						CacheFile: "/tmp/.cache",
					},
				},
			},
			[]string{`"cacheFile": "/tmp/.cache"`, `"name": "mycsp"`, `"namespace": "default"`},
		},
		"my new cstor pool": {
			"my new cstor pool",
			&CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "mynewcsp", Namespace: "openebs"},
				Spec: CStorPoolSpec{
					PoolSpec: CStorPoolAttr{
						CacheFile: "/tmp/.cache2",
					},
				},
			},
			[]string{`"cacheFile": "/tmp/.cache2"`, `"name": "mynewcsp"`, `"namespace": "openebs"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			jsonstr := stringer.JSONIndent(mock.context, mock.csp)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(jsonstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, jsonstr)
				}
			}
		})
	}
}

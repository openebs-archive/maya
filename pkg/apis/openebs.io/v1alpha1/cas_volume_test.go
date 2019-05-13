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

func TestCASVolumeString(t *testing.T) {
	tests := map[string]struct {
		volume              *CASVolume
		expectedStringParts []string
	}{
		"cas volume": {
			&CASVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "myvol", Namespace: "open"},
				Spec:       CASVolumeSpec{Capacity: "1G"},
			},
			[]string{"capacity: 1G", "name: myvol", "namespace: open"},
		},
		"cas volume 2": {
			&CASVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "mycoolvol", Namespace: "default"},
				Spec:       CASVolumeSpec{Capacity: "3G"},
			},
			[]string{"capacity: 3G", "name: mycoolvol", "namespace: default"},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.volume.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestCASVolumeJSONIndent(t *testing.T) {
	tests := map[string]struct {
		context             string
		volume              *CASVolume
		expectedStringParts []string
	}{
		"cas volume": {
			"my cas vol",
			&CASVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "myvol", Namespace: "open"},
				Spec:       CASVolumeSpec{Capacity: "1G"},
			},
			[]string{`"capacity": "1G"`, `"name": "myvol"`, `"namespace": "open"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			jsonstr := stringer.JSONIndent(mock.context, mock.volume)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(jsonstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, jsonstr)
				}
			}
		})
	}
}

func TestCASVolumeListString(t *testing.T) {
	tests := map[string]struct {
		volumelist          *CASVolumeList
		expectedStringParts []string
	}{
		"cas volume list": {
			&CASVolumeList{
				ObjectMeta: metav1.ObjectMeta{Name: "myvollist", Namespace: "open"},
				Items: []CASVolume{
					CASVolume{
						ObjectMeta: metav1.ObjectMeta{Name: "myvol", Namespace: "open"},
						Spec:       CASVolumeSpec{Capacity: "1G"},
					},
				},
			},
			[]string{"capacity: 1G", "name: myvol", "namespace: open"},
		},
		"cas volume list 2": {
			&CASVolumeList{
				ObjectMeta: metav1.ObjectMeta{Name: "myvollist2", Namespace: "default"},
				Items: []CASVolume{
					CASVolume{
						ObjectMeta: metav1.ObjectMeta{Name: "myvol2", Namespace: "default"},
						Spec:       CASVolumeSpec{Capacity: "2G"},
					},
				},
			},
			[]string{"capacity: 2G", "name: myvol2", "namespace: default"},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.volumelist.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestCASVolumeListJSONIndent(t *testing.T) {
	tests := map[string]struct {
		context             string
		volumelist          *CASVolumeList
		expectedStringParts []string
	}{
		"cas volume list": {
			"myvollist",
			&CASVolumeList{
				ObjectMeta: metav1.ObjectMeta{Name: "myvollist", Namespace: "open"},
				Items: []CASVolume{
					CASVolume{
						ObjectMeta: metav1.ObjectMeta{Name: "myvol", Namespace: "open"},
						Spec:       CASVolumeSpec{Capacity: "1G"},
					},
				},
			},
			[]string{`"capacity": "1G"`, `"name": "myvol"`, `"namespace": "open"`},
		},
		"cas volume list 2": {
			"myvollist2",
			&CASVolumeList{
				ObjectMeta: metav1.ObjectMeta{Name: "myvollist2", Namespace: "default"},
				Items: []CASVolume{
					CASVolume{
						ObjectMeta: metav1.ObjectMeta{Name: "myvol2", Namespace: "default"},
						Spec:       CASVolumeSpec{Capacity: "2G"},
					},
				},
			},
			[]string{`"capacity": "2G"`, `"name": "myvol2"`, `"namespace": "default"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := stringer.JSONIndent(mock.context, mock.volumelist)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

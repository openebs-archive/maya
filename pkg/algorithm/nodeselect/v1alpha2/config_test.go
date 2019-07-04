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

package v1alpha2

import (
	"reflect"
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		cspc *apis.CStorPoolCluster
		ns   string
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		// Test Case #1
		{
			name: "Getting instance of config#1",
			args:args{
				cspc:&apis.CStorPoolCluster{},
				ns:"openebs",
			},
			want:&Config{
				CSPC:&apis.CStorPoolCluster{},
				Namespace:"openebs",
			},
		},

		// Test Case #2
		{
			name: "Getting instance of config#2",
			args:args{
				cspc:&apis.CStorPoolCluster{},
				ns:"custom",
			},
			want:&Config{
				CSPC:&apis.CStorPoolCluster{},
				Namespace:"custom",
			},
		},

		// Test Case #3
		{
			name: "Getting instance of config#3",
			args:args{
				cspc:&apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cspc-1",
					},
				},
				ns:"custom",
			},
			want:&Config{
				CSPC:&apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cspc-1",
					},
				},
				Namespace:"custom",
			},
		},
	}
	for _, tt := range tests {
		// pin it
		tt:=tt
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(tt.args.cspc, tt.args.ns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

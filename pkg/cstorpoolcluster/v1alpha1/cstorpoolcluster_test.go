// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestCSPCList_Len(t *testing.T) {
	type fields struct {
		items []*CSPC
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "Test Case #1: Nil CSPC",
			fields: fields{
				items: nil,
			},
			want: 0,
		},

		{
			name: "Test Case #2: 2 CSPC items",
			fields: fields{
				items: fakeCSPCList([]string{"cspc-1", "cspc-2"}),
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &CSPCList{
				items: tt.fields.items,
			}
			if got := c.Len(); got != tt.want {
				t.Errorf("CSPCList.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCSPCList_ToAPIList(t *testing.T) {
	type fields struct {
		items []*CSPC
	}
	tests := []struct {
		name   string
		fields fields
		want   *apisv1alpha1.CStorPoolClusterList
	}{
		{
			name: "Test Case #1: 2 CSPC items",
			fields: fields{
				items: fakeCSPCList([]string{"cspc-1", "cspc-2"}),
			},
			want: fakeAPICSPCList([]string{"cspc-1", "cspc-2"}),
		},
		{
			name: "Test Case #2: Nil CSPC",
			fields: fields{
				items: nil,
			},
			want: &apisv1alpha1.CStorPoolClusterList{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &CSPCList{
				items: tt.fields.items,
			}
			if got := c.ToAPIList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CSPCList.ToAPIList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewForAPIObject(t *testing.T) {
	type args struct {
		obj  *apisv1alpha1.CStorPoolCluster
		opts []cspcBuildOption
	}
	tests := []struct {
		name string
		args args
		want *CSPC
	}{
		{
			name: "Test Case #1: Nil CSPC",
			args: args{
				obj: nil,
			},
			want: &CSPC{
				object: nil,
			},
		},

		{
			name: "Test Case #2: Non Nil CSPC",
			args: args{
				obj: &apisv1alpha1.CStorPoolCluster{},
			},
			want: &CSPC{
				object: &apisv1alpha1.CStorPoolCluster{},
			},
		},

		{
			name: "Test Case #3: Non Nil CSPC",
			args: args{
				obj: &apisv1alpha1.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cspc-1",
					},
				},
			},
			want: &CSPC{
				object: &apisv1alpha1.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cspc-1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := NewForAPIObject(tt.args.obj, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewForAPIObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCSPC_IsNil(t *testing.T) {
	type fields struct {
		object *apisv1alpha1.CStorPoolCluster
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Test Case #1: Nil CSPC",
			fields: fields{
				object: nil,
			},
			want: true,
		},

		{
			name: "Test Case #2: CSPC Not Nil",
			fields: fields{
				object: &apisv1alpha1.CStorPoolCluster{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &CSPC{
				object: tt.fields.object,
			}
			if got := c.IsNil(); got != tt.want {
				t.Errorf("CSPC.IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

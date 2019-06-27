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
	"errors"
	"reflect"
	"testing"

	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func fakeAPICSPCList(cspcNames []string) *apisv1alpha1.CStorPoolClusterList {
	if len(cspcNames) == 0 {
		return nil
	}

	list := &apisv1alpha1.CStorPoolClusterList{}
	for _, name := range cspcNames {
		cspc := apisv1alpha1.CStorPoolCluster{}
		cspc.SetName(name)
		list.Items = append(list.Items, cspc)
	}
	return list
}

func fakeCSPCList(cspcNames []string) []*CSPC {
	plist := []*CSPC{}
	for _, name := range cspcNames {
		cspc := apisv1alpha1.CStorPoolCluster{}
		cspc.SetName(name)
		plist = append(plist, &CSPC{&cspc})
	}
	return plist
}

func TestNewListBuilder(t *testing.T) {
	tests := []struct {
		name string
		want *ListBuilder
	}{
		{
			name: "Creating new list builder",
			want: &ListBuilder{list: &CSPCList{}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := NewListBuilder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewListBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListBuilderForAPIObjects(t *testing.T) {
	type args struct {
		cspcs *apisv1alpha1.CStorPoolClusterList
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "Test case #1: A CSPC API list having 2 items",
			args: args{
				cspcs: fakeAPICSPCList([]string{"cspc1", "cspc2"}),
			},
			wantLen: 2,
		},
		{
			name: "Test case #2: A CSPC API list having 1 items",
			args: args{
				cspcs: fakeAPICSPCList([]string{"cspc1"}),
			},
			wantLen: 1,
		},
		{
			name: "Test case #3: A CSPC API list having 5 items",
			args: args{
				cspcs: fakeAPICSPCList([]string{"cspc1", "cspc2", "cspc3", "cspc4", "cspc5"}),
			},
			wantLen: 5,
		},

		{
			name: "Test case #4: A CSPC API list having 0 items",
			args: args{
				cspcs: fakeAPICSPCList([]string{}),
			},
			wantLen: 0,
		},

		{
			name: "Test case #5: A nil CSPC",
			args: args{
				cspcs: nil,
			},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := ListBuilderForAPIObjects(tt.args.cspcs); got.list.Len() != tt.wantLen {
				t.Errorf("ListBuilderForAPIObjects Length = %v, wantLen %v", got, tt.wantLen)
			}
		})
	}
}

func TestListBuilderForObjects(t *testing.T) {
	type args struct {
		cspcs *CSPCList
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "Test Case #1: A CSPC List having 1 item",
			args: args{
				cspcs: &CSPCList{
					items: fakeCSPCList([]string{"cspc1"}),
				},
			},
			wantLen: 1,
		},

		{
			name: "Test Case #2: A CSPC List having 2 items",
			args: args{
				cspcs: &CSPCList{
					items: fakeCSPCList([]string{"cspc1", "cspc1"}),
				},
			},
			wantLen: 2,
		},

		{
			name: "Test Case #3: A CSPC List having 0 items",
			args: args{
				cspcs: &CSPCList{
					items: fakeCSPCList([]string{}),
				},
			},
			wantLen: 0,
		},

		{
			name: "Test Case #4: A nil CSPC list",
			args: args{
				cspcs: nil,
			},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := ListBuilderForObjects(tt.args.cspcs); got.list.Len() != tt.wantLen {
				t.Errorf("ListBuilderForObjects() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}

func TestListBuilder_List(t *testing.T) {
	type fields struct {
		list    *CSPCList
		filters PredicateList
		errs    []error
	}
	tests := []struct {
		name    string
		fields  fields
		wantLen int
		wantErr bool
	}{
		{
			name: "Test Case #1: CSPC List with 1 item",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1"}),
				},
			},
			wantLen: 1,
			wantErr: false,
		},

		{
			name: "Test Case #2: CSPC List with 2 items ",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1", "cspc2"}),
				},
			},
			wantLen: 2,
			wantErr: false,
		},

		{
			name: "Test Case #2: CSPC List with 2 items and error",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1", "cspc2"}),
				},
				errs: []error{errors.New("Mocked Error")},
			},
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &ListBuilder{
				list:    tt.fields.list,
				filters: tt.fields.filters,
				errs:    tt.fields.errs,
			}
			got, err := b.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBuilder.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Len() != tt.wantLen {
				t.Errorf("ListBuilder.List() Length = %v, want %v", got, tt.wantLen)
			}
		})
	}
}

func TestListBuilder_Len(t *testing.T) {
	type fields struct {
		list    *CSPCList
		filters PredicateList
		errs    []error
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			name: "Test Case #1",
			fields: fields{
				list: nil,
			},
			want:    0,
			wantErr: false,
		},

		{
			name: "Test Case #2",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1"}),
				},
			},
			want:    1,
			wantErr: false,
		},

		{
			name: "Test Case #3",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1"}),
				},
				errs: []error{errors.New("Mock Error")},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &ListBuilder{
				list:    tt.fields.list,
				filters: tt.fields.filters,
				errs:    tt.fields.errs,
			}
			got, err := b.Len()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBuilder.Len() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ListBuilder.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListBuilder_APIList(t *testing.T) {
	type fields struct {
		list    *CSPCList
		filters PredicateList
		errs    []error
	}
	tests := []struct {
		name    string
		fields  fields
		wantLen int
		wantErr bool
	}{
		{
			name: "Test Case #1",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1"}),
				},
			},
			wantLen: 1,
			wantErr: false,
		},

		{
			name: "Test Case #2",
			fields: fields{
				list: &CSPCList{
					items: fakeCSPCList([]string{"cspc1", "cspc1"}),
				},
			},
			wantLen: 2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &ListBuilder{
				list:    tt.fields.list,
				filters: tt.fields.filters,
				errs:    tt.fields.errs,
			}
			got, err := b.APIList()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBuilder.APIList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Items) != tt.wantLen {
				t.Errorf("ListBuilder.APIList() Length = %v, want %v", got, tt.wantLen)
			}
		})
	}
}

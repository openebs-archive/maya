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
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	poolspec "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolspecs"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	tests := []struct {
		name string
		want *Builder
	}{
		{
			name: "Test Case #1: Getting empty instance of builder object",
			want: &Builder{cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuilder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithName(t *testing.T) {
	type fields struct {
		cspc *CSPC
		errs []error
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Builder
		wantErr bool
	}{
		{
			name: "Test Case #1(Positive): Set the name field of the CSPC object",
			fields: fields{
				cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}},
			},
			args: args{
				name: "cstorpoolcluster-1",
			},
			want: &Builder{cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstorpoolcluster-1",
				},
			},
			},
			},
			wantErr: false,
		},

		{
			name: "Test Case #1(Negative): Set the name field of the CSPC object after injecting some error",
			fields: fields{
				cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}},
			},
			args: args{
				name: "",
			},
			want:    &Builder{cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				cspc: tt.fields.cspc,
				errs: tt.fields.errs,
			}
			if tt.wantErr == true {
				got := b.WithName((tt.args.name))
				if len(got.errs) != 1 {
					t.Error("Builder.WithName() = expected error but got none")
				}
			} else if got := b.WithName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithPoolSpecBuilder(t *testing.T) {
	type fields struct {
		cspc *CSPC
		errs []error
	}
	type args struct {
		poolSpecBuilder *poolspec.Builder
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        apisv1alpha1.PoolSpec
		injectError bool
	}{
		{
			name: "Test case #1(Positive): Set the pool specs ",
			fields: fields{
				cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}},
			},
			args: args{
				poolspec.NewBuilder(),
			},
			want:        apisv1alpha1.PoolSpec{},
			injectError: false,
		},

		{
			name: "Test case #1(Negative): Set the pool spec by injecting error in poolSpec builder",
			fields: fields{
				cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}},
			},
			args: args{
				poolspec.NewBuilder().AppendErrorToBuilder(errors.New("Mocked Error")),
			},
			want:        apisv1alpha1.PoolSpec{},
			injectError: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				cspc: tt.fields.cspc,
				errs: tt.fields.errs,
			}
			if tt.injectError {
				got := b.WithPoolSpecBuilder(tt.args.poolSpecBuilder)
				if len(got.errs) == 0 {
					t.Error("Expected error but got none")
				}
			} else if got := b.WithPoolSpecBuilder(tt.args.poolSpecBuilder); !reflect.DeepEqual(got.cspc.object.Spec.Pools[0], tt.want) {
				t.Errorf("Builder.WithPoolSpecBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	type fields struct {
		cspc *CSPC
		errs []error
	}
	tests := []struct {
		name    string
		fields  fields
		want    *CSPC
		wantErr bool
	}{
		{
			name: "Test Case #1(Positive): Getting the CSPC object",
			fields: fields{
				cspc: &CSPC{},
			},
			want:    &CSPC{},
			wantErr: false,
		},

		{
			name: "Test Case #1(Negative): Getting the CSPC object after injecting error",
			fields: fields{
				cspc: &CSPC{},
				errs: []error{errors.New("Mocked Error")},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				cspc: tt.fields.cspc,
				errs: tt.fields.errs,
			}
			got, err := b.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

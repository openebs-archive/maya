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

package upgrade

import (
	"errors"
	"reflect"
	"testing"

	"github.com/openebs/maya/pkg/version"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestIsCurrentVersionValid(t *testing.T) {
	type args struct {
		vd apis.VersionDetails
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive-test-1",
			args: args{
				vd: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "1.2.0",
					},
				},
			},
			want: true,
		},
		{
			name: "positive-test-2",
			args: args{
				vd: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "1.1.0",
					},
				},
			},
			want: true,
		},
		{
			name: "positive-test-3",
			args: args{
				vd: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "1.0.0",
					},
				},
			},
			want: true,
		},
		{
			name: "negative-test-1",
			args: args{
				vd: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "0.9.0",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt //pin it
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCurrentVersionValid(tt.args.vd); got != tt.want {
				t.Errorf("IsCurrentVersionValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDesiredVersionValid(t *testing.T) {
	type args struct {
		vd apis.VersionDetails
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive-test-1",
			args: args{
				vd: apis.VersionDetails{
					Desired: version.Current(),
				},
			},
			want: true,
		},
		{
			name: "negative-test-1",
			args: args{
				vd: apis.VersionDetails{
					Desired: "1.2.0",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt //pin it
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDesiredVersionValid(tt.args.vd); got != tt.want {
				t.Errorf("IsDesiredVersionValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	type args struct {
		vd apis.VersionDetails
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive-test-1",
			args: args{
				vd: apis.VersionDetails{
					Desired: "1.3.0",
					Status: apis.VersionStatus{
						Current: "1.2.0",
					},
				},
			},
			want: "1.2.0-1.3.0",
		},
	}
	for _, tt := range tests {
		tt := tt //pin it
		t.Run(tt.name, func(t *testing.T) {
			if got := Path(tt.args.vd); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetInProgressStatus(t *testing.T) {
	type args struct {
		vs apis.VersionStatus
	}
	tests := []struct {
		name string
		args args
		want apis.VersionStatus
	}{
		{
			name: "positive-test-1",
			args: args{
				vs: apis.VersionStatus{},
			},
			want: apis.VersionStatus{
				State: apis.ReconcileInProgress,
			},
		},
	}
	for _, tt := range tests {
		tt := tt //pin it
		t.Run(tt.name, func(t *testing.T) {
			got := SetInProgressStatus(tt.args.vs)
			// Exclude LastUpdateTime as it can never be same
			tt.want.LastUpdateTime = got.LastUpdateTime
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetInProgressStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetSuccessStatus(t *testing.T) {
	type args struct {
		vs apis.VersionStatus
	}
	tests := []struct {
		name string
		args args
		want apis.VersionStatus
	}{
		{
			name: "positive-test-1",
			args: args{
				vs: apis.VersionStatus{},
			},
			want: apis.VersionStatus{
				Current: version.Current(),
				State:   apis.ReconcileComplete,
			},
		},
	}
	for _, tt := range tests {
		tt := tt //pin it
		t.Run(tt.name, func(t *testing.T) {
			got := SetSuccessStatus(tt.args.vs)
			// Exclude LastUpdateTime as it can never be same
			tt.want.LastUpdateTime = got.LastUpdateTime
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetSuccessStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetErrorStatus(t *testing.T) {
	type args struct {
		vs  apis.VersionStatus
		msg string
		err error
	}
	tests := []struct {
		name string
		args args
		want apis.VersionStatus
	}{
		{
			name: "positive-test-1",
			args: args{
				vs: apis.VersionStatus{
					State: apis.ReconcileInProgress,
				},
				msg: "failed to reconcile resource version",
				err: errors.New("invalid current version 0.9.0"),
			},
			want: apis.VersionStatus{
				State:   apis.ReconcileInProgress,
				Message: "failed to reconcile resource version",
				Reason:  "invalid current version 0.9.0",
			},
		},
	}
	for _, tt := range tests {
		tt := tt //pin it
		t.Run(tt.name, func(t *testing.T) {
			got := SetErrorStatus(tt.args.vs, tt.args.msg, tt.args.err)
			// Exclude LastUpdateTime as it can never be same
			tt.want.LastUpdateTime = got.LastUpdateTime
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetErrorStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func Test_getBaseImage(t *testing.T) {
	type args struct {
		deployObj *appsv1.Deployment
		name      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "Without proxy",
			want:    "quay.io/openebs/cstor-pool",
			wantErr: false,
			args: args{
				deployObj: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name:  "cstor-pool",
										Image: "quay.io/openebs/cstor-pool:1.4.0",
									},
								},
							},
						},
					},
				},
				name: "cstor-pool",
			},
		},
		{
			name:    "With proxy",
			want:    "fsdepot.evry.com:8085/openebs/cstor-pool",
			wantErr: false,
			args: args{
				deployObj: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name:  "cstor-pool",
										Image: "fsdepot.evry.com:8085/openebs/cstor-pool:1.4.0",
									},
								},
							},
						},
					},
				},
				name: "cstor-pool",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBaseImage(tt.args.deployObj, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBaseImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getBaseImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

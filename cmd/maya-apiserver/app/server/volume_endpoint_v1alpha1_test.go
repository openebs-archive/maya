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

package server

import (
	"testing"

	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsNotFound(t *testing.T) {
	tests := map[string]struct {
		err              error
		expectIsNotFound bool
	}{
		"nil error": {
			nil,
			false,
		},
		"nil error when wrapped": {
			errors.Wrap(nil, "wrapper thing"),
			false,
		},
		"any error": {
			errors.New("any error"),
			false,
		},
		"any error when wrapped": {
			errors.Wrap(errors.New("any error"), "wrapper thing"),
			false,
		},
		"any error when wrapped twice": {
			errors.Wrap(errors.Wrap(errors.New("any error"), "wrapper thing"), "wrap again"),
			false,
		},
		"template isnotfound error": {
			&templatefuncs.NotFoundError{
				ErrMsg: "catch me",
			},
			true,
		},
		"template isnotfound error when wrapped": {
			errors.Wrap(
				&templatefuncs.NotFoundError{
					ErrMsg: "catch me",
				},
				"wrapper thing",
			),
			true,
		},
		"template isnotfound error when wrapped twice": {
			errors.Wrap(errors.Wrap(
				&templatefuncs.NotFoundError{
					ErrMsg: "catch me",
				},
				"wrapper thing",
			), "wrap again"),
			true,
		},
		"k8s isnotfound error": {
			&k8serrors.StatusError{metav1.Status{
				Reason: metav1.StatusReasonNotFound,
			}},
			true,
		},
		"k8s isnotfound error when wrapped": {
			errors.Wrap(
				&k8serrors.StatusError{metav1.Status{
					Reason: metav1.StatusReasonNotFound,
				}},
				"wrapper thing",
			),
			true,
		},
		"k8s isnotfound error when wrapped twice": {
			errors.Wrap(errors.Wrap(
				&k8serrors.StatusError{metav1.Status{
					Reason: metav1.StatusReasonNotFound,
				}},
				"wrapper thing",
			), "wrap again"),
			true,
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			actual := isNotFound(mock.err)
			if actual != mock.expectIsNotFound {
				t.Errorf("test '%s' failed: expected isNotFound error as '%t' got '%t'", name, mock.expectIsNotFound, actual)
			}
		})
	}
}

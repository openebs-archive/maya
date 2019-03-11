// Copyright © 2018-2019 The OpenEBS Authors
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

package volume

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	v1_storage "k8s.io/api/storage/v1"
)

func TestGetCreateCASTemplate(t *testing.T) {
	sc := &v1_storage.StorageClass{}
	sc.Annotations = make(map[string]string)
	tests := map[string]struct {
		scCreateCASAnnotation string
		scCASTypeAnnotation   string
		defaultCasType        string
		envJivaCAST           string
		envCStorCAST          string
		expectedCAST          string
	}{
		"CAST annotation is present": {
			"cast-create-from-annotation",
			"",
			"",
			"",
			"",
			"cast-create-from-annotation",
		},
		"CAST annotation is absent/empty and cas type is cstor, defaultCasType is jiva": {
			"",
			"cstor",
			"jiva",
			"",
			"cast-cstor-create-from-env",
			"cast-cstor-create-from-env",
		},
		"CAST annotation is absent/empty and cas type is jiva": {
			"",
			"jiva",
			"",
			"cast-jiva-create-from-env",
			"",
			"cast-jiva-create-from-env",
		},
		"CAST annotation is absent/empty and cas type is missing, defaultCasType is cstor": {
			"",
			"",
			"cstor",
			"",
			"cast-cstor-create-from-env",
			"cast-cstor-create-from-env",
		},
		"CAST annotation is absent/empty and cas type unknown": {
			"",
			"unknown",
			"",
			"cast-jiva-create-from-env",
			"cast-cstor-create-from-env",
			"",
		},
	}

	defer func() {
		os.Unsetenv(string(menv.CASTemplateToCreateCStorVolumeENVK))
		os.Unsetenv(string(menv.CASTemplateToCreateJivaVolumeENVK))
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeCreate)] = test.scCreateCASAnnotation
			sc.Annotations[string(v1alpha1.CASTypeKey)] = test.scCASTypeAnnotation
			os.Setenv(string(menv.CASTemplateToCreateCStorVolumeENVK), test.envCStorCAST)
			os.Setenv(string(menv.CASTemplateToCreateJivaVolumeENVK), test.envJivaCAST)

			castName := getCreateCASTemplate(test.defaultCasType, sc)

			if castName != test.expectedCAST {
				t.Fatalf("unexpected cast name, wanted %q got %q", test.expectedCAST, castName)
			}
		})
	}
}

func TestGetReadCASTemplate(t *testing.T) {
	sc := &v1_storage.StorageClass{}
	sc.Annotations = make(map[string]string)
	tests := map[string]struct {
		scReadCASAnnotation string
		scCASTypeAnnotation string
		defaultCasType      string
		envJivaCAST         string
		envCStorCAST        string
		expectedCAST        string
	}{
		"CAST annotation is present": {
			"cast-read-from-annotation",
			"",
			"",
			"",
			"",
			"cast-read-from-annotation",
		},
		"CAST annotation is absent/empty and cas type is cstor, defaultCasType is jiva": {
			"",
			"cstor",
			"jiva",
			"",
			"cast-cstor-read-from-env",
			"cast-cstor-read-from-env",
		},
		"CAST annotation is absent/empty and cas type is jiva": {
			"",
			"jiva",
			"",
			"cast-jiva-read-from-env",
			"",
			"cast-jiva-read-from-env",
		},
		"CAST annotation is absent/empty and cas type is missing, defaultCasType is cstor": {
			"",
			"",
			"cstor",
			"",
			"cast-cstor-read-from-env",
			"cast-cstor-read-from-env",
		},
		"CAST annotation is absent/empty and cas type unknown": {
			"",
			"unknown",
			"",
			"cast-jiva-read-from-env",
			"cast-cstor-read-from-env",
			"",
		},
	}

	defer func() {
		os.Unsetenv(string(menv.CASTemplateToCreateCStorVolumeENVK))
		os.Unsetenv(string(menv.CASTemplateToCreateJivaVolumeENVK))
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeRead)] = test.scReadCASAnnotation
			sc.Annotations[string(v1alpha1.CASTypeKey)] = test.scCASTypeAnnotation
			os.Setenv(string(menv.CASTemplateToReadCStorVolumeENVK), test.envCStorCAST)
			os.Setenv(string(menv.CASTemplateToReadJivaVolumeENVK), test.envJivaCAST)

			castName := getReadCASTemplate(test.defaultCasType, sc)

			if castName != test.expectedCAST {
				t.Fatalf("unexpected cast name, wanted %q got %q", test.expectedCAST, castName)
			}
		})
	}
}

func TestGetDeleteCASTemplate(t *testing.T) {
	sc := &v1_storage.StorageClass{}
	sc.Annotations = make(map[string]string)
	tests := map[string]struct {
		scDeleteCASAnnotation string
		scCASTypeAnnotation   string
		defaultCasType        string
		envJivaCAST           string
		envCStorCAST          string
		expectedCAST          string
	}{
		"CAST annotation is present": {
			"cast-delete-from-annotation",
			"",
			"",
			"",
			"",
			"cast-delete-from-annotation",
		},
		"CAST annotation is absent/empty and cas type is cstor, defaultCasType is jiva": {
			"",
			"cstor",
			"jiva",
			"",
			"cast-cstor-delete-from-env",
			"cast-cstor-delete-from-env",
		},
		"CAST annotation is absent/empty and cas type is jiva": {
			"",
			"jiva",
			"",
			"cast-jiva-read-from-env",
			"",
			"cast-jiva-read-from-env",
		},
		"CAST annotation is absent/empty and cas type is missing, defaultCasType is cstor": {
			"",
			"",
			"cstor",
			"",
			"cast-cstor-delete-from-env",
			"cast-cstor-delete-from-env",
		},
		"CAST annotation is absent/empty and cas type unknown": {
			"",
			"unknown",
			"",
			"cast-jiva-delete-from-env",
			"cast-cstor-delete-from-env",
			"",
		},
	}

	defer func() {
		os.Unsetenv(string(menv.CASTemplateToCreateCStorVolumeENVK))
		os.Unsetenv(string(menv.CASTemplateToCreateJivaVolumeENVK))
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeDelete)] = test.scDeleteCASAnnotation
			sc.Annotations[string(v1alpha1.CASTypeKey)] = test.scCASTypeAnnotation
			os.Setenv(string(menv.CASTemplateToDeleteCStorVolumeENVK), test.envCStorCAST)
			os.Setenv(string(menv.CASTemplateToDeleteJivaVolumeENVK), test.envJivaCAST)

			castName := getDeleteCASTemplate(test.defaultCasType, sc)

			if castName != test.expectedCAST {
				t.Fatalf("unexpected cast name, wanted %q got %q", test.expectedCAST, castName)
			}
		})
	}
}

func TestgetString(t *testing.T) {
	tests := map[string]struct {
		testString    string
		expectedValue string
		expectedError error
	}{
		"Interms of G": {
			testString:    "10G",
			expectedValue: "G",
			expectedError: nil,
		},
		"Interms of Zi": {
			testString:    "5Zi",
			expectedValue: "Zi",
			expectedError: nil,
		},
		"Interms of T": {
			testString:    "9T",
			expectedValue: "T",
			expectedError: nil,
		},
		"Interms of K": {
			testString:    "9.04K",
			expectedValue: "K",
			expectedError: nil,
		},
		"Invalid Unit": {
			testString:    "9.04S",
			expectedValue: "",
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := getString(test.testString)
			if val != test.expectedValue && err == nil {
				t.Logf("actual string: %s", val)
				t.Logf("expected string: %s", test.expectedValue)
				t.Error("unexpected rounded value")
			}
			if !reflect.DeepEqual(err, test.expectedError) {
				t.Errorf("Test case: %s Expected error: %v but got: %v", name, test.expectedError, err)
			}
		})
	}
}

func TestgetInt(t *testing.T) {
	tests := map[string]struct {
		testString    string
		expectedValue uint64
		expectedError error
	}{
		"Interms of G": {
			testString:    "10G",
			expectedValue: uint64(10),
			expectedError: nil,
		},
		"Interms of Zi": {
			testString:    "5Zi",
			expectedValue: uint64(5),
			expectedError: nil,
		},
		"Interms of T": {
			testString:    "9T",
			expectedValue: uint64(9),
			expectedError: nil,
		},
		"Interms of K": {
			testString:    "3K",
			expectedValue: uint64(3),
			expectedError: nil,
		},
		"Invalid Unit": {
			testString:    "Mi",
			expectedValue: uint64(64),
			expectedError: fmt.Errorf("No number found"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := getInt(test.testString)
			if val != test.expectedValue && err == nil {
				t.Logf("actual string: %d", val)
				t.Logf("expected string: %d", test.expectedValue)
				t.Error("unexpected rounded value")
			}
			if !reflect.DeepEqual(err, test.expectedError) {
				t.Errorf("Test case Name: '%s' Expected error: '%v' but got: '%v'", name, test.expectedError, err)
			}
		})
	}
}

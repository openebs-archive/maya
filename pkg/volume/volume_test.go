package volume

import (
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	v1_storage "k8s.io/api/storage/v1"
	"os"
	"testing"
)

func TestGetCreateCASTemplate(t *testing.T) {
	sc := &v1_storage.StorageClass{}
	sc.Annotations = make(map[string]string)
	tests := map[string]struct {
		scCreateCASAnnotation string
		scCASTypeAnnotation   string
		envJivaCAST           string
		envCStorCAST          string
		expectedCAST          string
	}{
		"CAST annotation is present": {
			"cast-create-from-annotation",
			"",
			"",
			"",
			"cast-create-from-annotation",
		},
		"CAST annotation is absent/empty and cas type is cstor": {
			"",
			"cstor",
			"",
			"cast-cstor-create-from-env",
			"cast-cstor-create-from-env",
		},
		"CAST annotation is absent/empty and cas type is jiva": {
			"",
			"jiva",
			"cast-jiva-create-from-env",
			"",
			"cast-jiva-create-from-env",
		},
		"CAST annotation is absent/empty and cas type unknown": {
			"",
			"unknown",
			"cast-jiva-create-from-env",
			"cast-cstor-create-from-env",
			"",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeCreate)] = test.scCreateCASAnnotation
			sc.Annotations[string(v1alpha1.CASTypeKey)] = test.scCASTypeAnnotation
			os.Setenv(string(menv.CASTemplateToCreateCStorVolumeENVK), test.envCStorCAST)
			os.Setenv(string(menv.CASTemplateToCreateJivaVolumeENVK), test.envJivaCAST)

			castName := getCreateCASTemplate(sc)

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
		envJivaCAST         string
		envCStorCAST        string
		expectedCAST        string
	}{
		"CAST annotation is present": {
			"cast-read-from-annotation",
			"",
			"",
			"",
			"cast-read-from-annotation",
		},
		"CAST annotation is absent/empty and cas type is cstor": {
			"",
			"cstor",
			"",
			"cast-cstor-read-from-env",
			"cast-cstor-read-from-env",
		},
		"CAST annotation is absent/empty and cas type is jiva": {
			"",
			"jiva",
			"cast-jiva-read-from-env",
			"",
			"cast-jiva-read-from-env",
		},
		"CAST annotation is absent/empty and cas type unknown": {
			"",
			"unknown",
			"cast-jiva-read-from-env",
			"cast-cstor-read-from-env",
			"",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeRead)] = test.scReadCASAnnotation
			sc.Annotations[string(v1alpha1.CASTypeKey)] = test.scCASTypeAnnotation
			os.Setenv(string(menv.CASTemplateToReadCStorVolumeENVK), test.envCStorCAST)
			os.Setenv(string(menv.CASTemplateToReadJivaVolumeENVK), test.envJivaCAST)

			castName := getReadCASTemplate(sc)

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
		envJivaCAST           string
		envCStorCAST          string
		expectedCAST          string
	}{
		"CAST annotation is present": {
			"cast-read-from-annotation",
			"",
			"",
			"",
			"cast-read-from-annotation",
		},
		"CAST annotation is absent/empty and cas type is cstor": {
			"",
			"cstor",
			"",
			"cast-cstor-read-from-env",
			"cast-cstor-read-from-env",
		},
		"CAST annotation is absent/empty and cas type is jiva": {
			"",
			"jiva",
			"cast-jiva-read-from-env",
			"",
			"cast-jiva-read-from-env",
		},
		"CAST annotation is absent/empty and cas type unknown": {
			"",
			"unknown",
			"cast-jiva-read-from-env",
			"cast-cstor-read-from-env",
			"",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeDelete)] = test.scDeleteCASAnnotation
			sc.Annotations[string(v1alpha1.CASTypeKey)] = test.scCASTypeAnnotation
			os.Setenv(string(menv.CASTemplateToDeleteCStorVolumeENVK), test.envCStorCAST)
			os.Setenv(string(menv.CASTemplateToDeleteJivaVolumeENVK), test.envJivaCAST)

			castName := getDeleteCASTemplate(sc)

			if castName != test.expectedCAST {
				t.Fatalf("unexpected cast name, wanted %q got %q", test.expectedCAST, castName)
			}
		})
	}
}

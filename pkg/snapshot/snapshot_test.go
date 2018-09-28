package snapshot

import (
	"os"
	"testing"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
)

func Test_getCreateCASTemplate(t *testing.T) {
	os.Setenv(string(menv.CASTemplateToCreateCStorSnapshotENVK), "cstor-cast")
	os.Setenv(string(menv.CASTemplateToCreateJivaSnapshotENVK), "jiva-cast")
	defer os.Unsetenv(string(menv.CASTemplateToCreateCStorSnapshotENVK))
	defer os.Unsetenv(string(menv.CASTemplateToCreateJivaSnapshotENVK))
	tests := map[string]struct {
		casType      string
		wantCastName string
	}{
		"casType is 'cstor'": {
			casType:      "cstor",
			wantCastName: "cstor-cast",
		},
		"casType is 'jiva'": {
			casType:      "jiva",
			wantCastName: "jiva-cast",
		},
		"casType is empty": {
			casType:      "",
			wantCastName: "jiva-cast",
		},
		"casType is 'unknown'": {
			casType:      "unknown",
			wantCastName: "",
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			if gotCastName := getCreateCASTemplate(mock.casType); gotCastName != mock.wantCastName {
				t.Errorf("getCreateCASTemplate() = %v, want %v", gotCastName, mock.wantCastName)
			}
		})
	}
}

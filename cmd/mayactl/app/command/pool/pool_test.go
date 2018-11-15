package pool

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCmdPool(t *testing.T) {
	tests := map[string]*struct {
		expectedCmd *cobra.Command
	}{
		"NewCmdVolumeStats": {
			expectedCmd: &cobra.Command{
				Use:   "pool",
				Short: "Provides operations related to a storage pool",
				Long:  poolCommandHelpText,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewCmdPool()
			if (got.Use != tt.expectedCmd.Use) || (got.Short != tt.expectedCmd.Short) || (got.Long != tt.expectedCmd.Long) || (got.Example != tt.expectedCmd.Example) {
				t.Fatalf("TestName: %v | processStats() => Got: %v | Want: %v \n", name, got, tt.expectedCmd)
			}
		})
	}
}

// returns true when both errors are true or else returns false
func checkErr(err1, err2 error) bool {
	if (err1 != nil && err2 == nil) || (err1 == nil && err2 != nil) || (err1 != nil && err2 != nil && err1.Error() != err2.Error()) {
		return false
	}
	return true
}

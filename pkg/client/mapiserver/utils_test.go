package mapiserver

import (
	"os"
	"testing"
)

func TestInitialize(t *testing.T) {
	tests := map[string]struct {
		key   string
		value string
	}{
		"MAPI_ADDRSet":    {"MAPI_ADDR", "127.0.0.1"},
		"MAPI_ADDRNotSet": {"MAPI_ADDR", ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(tt.key, tt.value)
			defer os.Unsetenv(tt.key)
			Initialize()
		})
	}
}

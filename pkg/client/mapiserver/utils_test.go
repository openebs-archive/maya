package mapiserver

import (
	"os"
	"reflect"
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

func TestGetURL(t *testing.T) {
	cases := map[string]*struct {
		addr           string
		envaddr        string
		expectedoutput string
	}{
		"Environment variable set": {
			envaddr:        "192.168.0.2",
			expectedoutput: "192.168.0.2",
		},
		"Environment vaiable not set": {
			addr:           "192.168.0.1",
			expectedoutput: "http://192.168.0.1:5656",
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if len(tt.envaddr) > 0 {
				os.Setenv("MAPI_ADDR", tt.envaddr)

			} else {
				MAPIAddr = tt.addr
			}

			got := GetURL()
			os.Unsetenv("MAPI_ADDR")
			MAPIAddr = ""
			if !reflect.DeepEqual(got, tt.expectedoutput) {
				t.Fatalf("GetURL => got %v, want %v ", got, tt.expectedoutput)
			}

		})
	}
}

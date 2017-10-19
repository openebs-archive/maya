package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMayaConfig_Parse(t *testing.T) {
	cases := []struct {
		File   string
		Result *MayaConfig
		Err    bool
	}{
		{
			"dummy_mayaserver_config.hcl",
			&MayaConfig{
				Region:      "BANG-EAST",
				Datacenter:  "dc2",
				NodeName:    "my-vsm",
				DataDir:     "/tmp/mayaserver",
				LogLevel:    "ERR",
				BindAddr:    "192.168.0.1",
				EnableDebug: true,
				Ports: &Ports{
					HTTP: 1234,
				},
				Addresses: &Addresses{
					HTTP: "127.0.0.1",
				},
				AdvertiseAddrs: &AdvertiseAddrs{},
				LeaveOnInt:     true,
				LeaveOnTerm:    true,
				EnableSyslog:   true,
				SyslogFacility: "LOCAL1",
				HTTPAPIResponseHeaders: map[string]string{
					"Access-Control-Allow-Origin": "*",
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Logf("Testing parse: %s", tc.File)

		path, err := filepath.Abs(filepath.Join("../mockit", tc.File))
		if err != nil {
			t.Fatalf("file: %s\n\n%s", tc.File, err)
			continue
		}

		actual, err := ParseMayaConfigFile(path)
		if (err != nil) != tc.Err {
			t.Fatalf("file: %s\n\n%s", tc.File, err)
			continue
		}

		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("file: %s\nactual:%q\n\nexpected:%q", tc.File, actual, tc.Result)
		}
	}
}

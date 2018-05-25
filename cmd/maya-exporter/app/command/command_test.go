package command

import (
	"errors"
	"reflect"
	"testing"
	"net"
	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Short: "Collect metrics from OpenEBS volumes",
		Long: `maya-exporter can be used to monitor openebs volumes and pools.
It can be deployed alongside the openebs volume or pool containers as sidecars.`,
		Example: `maya-exporter -a=http://localhost:8001 -c=:9500 -m=/metrics`,
	}
)

func TestRegisterJivaStatsExporter(t *testing.T) {
	cases := map[string]struct {
		option *VolumeExporterOptions
		output error
	}{
		"ValidURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "http://localhost:9501",
			},
			output: nil,
		},
		"InvalidURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "localhost",
			},
			output: errors.New("Error in parsing the URI"),
		},
		"EmptyURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "",
			},
			output: errors.New("Error in parsing the URI"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := tt.option.RegisterJivaStatsExporter()
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("RegisterJivaStatsExporter() => [%v], want [%v]", got, tt.output)
			}
		})
	}

}

func TestRegisterCstorStatsExporter(t *testing.T) {
	start := make(chan struct{})
	cases := map[string]struct {
		option   *VolumeExporterOptions
		sockPath string
		output   error
	}{
		"ValidSocketPath": {
			option: &VolumeExporterOptions{
			},
			sockPath: "anysocketpath",
			output:   nil,
		},
		"InvalidSocketPath": {
			option: &VolumeExporterOptions{
			},
			sockPath: "/var/sock",
			output:   errors.New("Error in initiating the connection"),
		},
		"EmptyPath": {
			option: &VolumeExporterOptions{
			},

			output: errors.New("Error in initiating the connection"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			go func() {
				<-start
	            ln, err := net.Listen("unix", "anysocketpath")
	            if err != nil {
		           t.Fatal(err)
	            }
	            for {
		           _, err := ln.Accept()
		           if err != nil {
				       t.Fatal("Accept error: ", err)
		           }
                }
			}()
			go func() {
				got := tt.option.RegisterCstorStatsExporter(tt.sockPath)
			    if !reflect.DeepEqual(got, tt.output) {
				    t.Fatalf("RegisterCstorStatsExporter() => [%v], want [%v]", got, tt.output)
				}
				close(start)
			}()
		})
	}

}

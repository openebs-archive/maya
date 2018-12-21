package collector

import (
	"reflect"
	"testing"

	v1 "github.com/openebs/maya/pkg/apis/openebs.io/stats"
)

func TestNewResponse(t *testing.T) {
	cases := map[string]struct {
		response string
		output   v1.VolumeStats
	}{
		"[Success]Unmarshal Response into Metrics struct": {
			response: JSONFormatedResponse,
			output: v1.VolumeStats{
				Size:                 "10737418240",
				Iqn:                  "iqn.2017-08.OpenEBS.cstor:vol1",
				Writes:               "0",
				Reads:                "0",
				TotalReadBytes:       "0",
				TotalWriteBytes:      "0",
				UsedLogicalBlocks:    "19",
				SectorSize:           "512",
				TotalReadBlockCount:  "12",
				TotalWriteBlockCount: "15",
				TotalReadTime:        "13",
				TotalWriteTime:       "132",
				UpTime:               "20",
				ReplicaCounter:       "3",
				RevisionCounter:      "1000",
				Replicas: []v1.Replica{
					{
						Address: "tcp://172.18.0.3:9502",
						Mode:    "DEGRADED",
					},
					{
						Address: "tcp://172.18.0.4:9502",
						Mode:    "HEALTHY",
					},
					{
						Address: "tcp://172.18.0.5:9502",
						Mode:    "HEALTHY",
					},
				},
			},
		},
		"[Failure]Unmarshal Response returns empty Metrics": {
			response: ImproperJSONFormatedResponse,
			output:   v1.VolumeStats{},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := newResponse(tt.response)
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("unmarshal(%v) : expected %v, got %v", tt.response, tt.output, got)
			}
		})
	}
}

func TestSplitter(t *testing.T) {
	cases := map[string]struct {
		response         string
		splittedResponse string
	}{
		"[Success] If response is as expected": {
			response:         CstorResponse,
			splittedResponse: SplittedResponse,
		},
		"[Failure] If response is not as expected, splitter should return nil string": {
			response:         NilCstorResponse,
			splittedResponse: "",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if got := splitter(tt.response); !reflect.DeepEqual(got, tt.splittedResponse) {
				t.Fatalf("splitter(%v) => expected %v, got %v", tt.response, tt.splittedResponse, got)
			}
		})
	}

}

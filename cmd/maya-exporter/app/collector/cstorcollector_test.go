package collector

import (
	"testing"
)

var (
	resp1 string = "iSCSI Target Controller version istgt:0.5.20121028:20:26:33:May 21 2018 on  from\r\nIOSTATS  IQN=iqn.2017-08.OpenEBS.cstor:vol1 Blockcount=20971520 Blocklength=512 Writes=0 Reads=0 TotalReadBytes=0 TotalWriteBytes=0 Size=10737418240\r\nOK IOSTATS\r\n"
	resp2 string = "IOSTATS  IQN=iqn.2017-08.OpenEBS.cstor:vol1 Blockcount=20971520 Blocklength=512 Writes=0 Reads=0 TotalReadBytes=0 TotalWriteBytes=0 Size=10737418240\r\nOK IOSTATS\r\n"
)

func TestSplit(t *testing.T) {
	cases := map[string]struct {
		response string
		output   map[string]string
	}{
		"Response with header": {
			response: resp1,
			output: map[string]string{
				"TotalWriteBytes": "0",
				"Size":            "10737418240",
				"IQN":             "iqn.2017-08.OpenEBS.cstor:vol1",
				"Blockcount":      "20971520",
				"Blocklength":     "512",
				"Writes":          "0",
				"Reads":           "0",
				"TotalReadBytes":  "0",
			},
		},
		"Response without header": {
			response: resp2,
			output: map[string]string{
				"TotalWriteBytes": "0",
				"Size":            "10737418240",
				"IQN":             "iqn.2017-08.OpenEBS.cstor:vol1",
				"Blockcount":      "20971520",
				"Blocklength":     "512",
				"Writes":          "0",
				"Reads":           "0",
				"TotalReadBytes":  "0",
			},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := split(tt.response)
			if len(got) != len(tt.output) {
				t.Fatalf("split(%v) : expected %v, got %v", tt.response, got, tt.output)
			}
		})
	}
}

package collector

import (
	"encoding/json"
	"strings"

	"github.com/golang/glog"
	v1 "github.com/openebs/maya/pkg/apis/openebs.io/stats"
)

type connErr struct {
	err error
}

func (e *connErr) Error() string {
	return e.err.Error()
}

// removeItem removes the string passed as argument from the slice
func removeItem(slice []string, str string) []string {
	for index, value := range slice {
		if value == str {
			slice = append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}

// newResponse unmarshal the JSON into Response instances.
func newResponse(result string) (v1.VolumeStats, error) {
	metrics := v1.VolumeStats{}
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		glog.Error("Error in unmarshalling, found error: ", err)
		return metrics, err
	}
	return metrics, nil
}

// splitter extracts the JSON from the response :
// "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\",
//	\"Writes\": \"0\", \"Reads\": \"0\", \"TotalWriteBytes\": \"0\",
//  \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
func splitter(resp string) string {
	var result []string
	result = strings.Split(resp, EOF)
	result = removeItem(result, Footer)
	if len(result[0]) == 0 {
		return ""
	}
	res := strings.TrimPrefix(result[0], Command+"  ")
	return res
}

package version

import (
	"net/http/httptest"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	response = `[{"tag_name": "0.4.4"}]`
)

func TestGetLatestRelease(t *testing.T) {
	cases := map[string]struct {
		fakeHandler utiltesting.FakeHandler
		output      string
		err         error
	}{
		"NoResponse": {
			fakeHandler: utiltesting.FakeHandler{
				// ResponseBody: string(response),
				T: t,
			},
			output: "",
		},
		"Response": {
			fakeHandler: utiltesting.FakeHandler{
				ResponseBody: string(response),
				T:            t,
			},
			output: "0.4.4",
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			GitAPI = server.URL
			got, _ := GetLatestRelease()
			if got != tt.output {
				t.Errorf("Actual - %s Expected - %s", got, tt.output)
			}
			defer server.Close()
		})
	}
}

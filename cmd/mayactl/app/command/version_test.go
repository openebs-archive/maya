package command

import (
	"net/http/httptest"
	"testing"

	"github.com/openebs/maya/pkg/version"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	response2 = `[{"tag_name": "0.5.4"}]`
)

func TestCheckLatestVersion(t *testing.T) {
	cases := map[string]struct {
		installed string
		behaviour string
	}{
		"OnLatestVersion":         {"v0.5.4", "pos"},
		"NewerVersionAvailable":   {"v0.5.6", "pos"},
		"NewerVersionAvailable-1": {"v0.4.1", "pos"},
		"VersionContainingString": {"v0.5.4-test", "pos"},
		"EmptyString":             {"", "neg"},
		"InvalidString":           {"0.5.7", "neg"},
	}

	fakeHandler :=
		utiltesting.FakeHandler{
			StatusCode:   200,
			ResponseBody: string(response2),
		}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&fakeHandler)
			version.GitAPIAddr = server.URL
			err := checkLatestVersion(c.installed)
			if c.behaviour == "pos" && err != nil {
				t.Errorf("TestName: '%s' ExpectedErr: 'nil' ActualErr: '%s'", name, err.Error())
			}
			if c.behaviour == "neg" && err == nil {
				t.Errorf("TestName: '%s' ExpectedErr: '%s' ActualErr: 'nil'", name, err.Error())
			}
			defer server.Close()
		})
	}

}

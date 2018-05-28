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
	cases := []struct {
		installed string
		behaviour string
	}{
		{"v0.5.4", "pos"},
		{"v0.5.6", "pos"},
		{"v0.4.1", "pos"},
		{"", "neg"},
		{"0.5.7", "neg"},
	}

	fakeHandler :=
		utiltesting.FakeHandler{
			ResponseBody: string(response2),
		}

	for i, c := range cases {
		server := httptest.NewServer(&fakeHandler)
		version.GitAPI = server.URL
		err := checkLatestVersion(c.installed)
		if c.behaviour == "pos" && err != nil {
			t.Errorf("TestCase: '%d' ExpectedErr: 'nil' ActualErr: '%s'", i, err.Error())
		}
		if c.behaviour == "neg" && err == nil {
			t.Errorf("TestCase: '%d' ExpectedErr: '%s' ActualErr: 'nil'", i, err.Error())
		}
		defer server.Close()
	}

}

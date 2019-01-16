package collector

import (
	"testing"

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
)

type fakeVol struct {
}

func (f *fakeVol) get() (v1.VolumeStats, error) {
	return v1.VolumeStats{}, nil
}
func (f *fakeVol) parse(volStats v1.VolumeStats, metrics *metrics) stats {
	return stats{}
}

func TestCasType(t *testing.T) {
	cases := map[string]struct {
		vol Volume
		cas string
	}{
		"cas type is jiva": {
			vol: new(jiva),
			cas: "jiva",
		},
		"cas type is cstor": {
			vol: new(cstor),
			cas: "cstor",
		},
		"cas type is fakeVol": {
			vol: new(fakeVol),
			cas: "",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := casType(tt.vol)
			if got != tt.cas {
				t.Fatalf("casType(%v): expected %v, got %v", tt.vol, tt.cas, got)
			}
		})
	}
}

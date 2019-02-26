package v1alpha1

import (
	"testing"
)

func TestCstorVolumeResize(t *testing.T) {
	tests := map[string]struct {
		ip       string
		volName  string
		capacity string
		isErr    bool
		errMsg   string
	}{
		"test 101": {"", "vol1", "", true, "failed to resize the cstor volume: missing ip address"},
		"test 102": {"0.0.0.0", "", "", true, "failed to resize the cstor volume: missing volume name"},
		"test 103": {"0.0.0.0", "pvc-21312-321312-321321-31231", "", true, "failed to resize the cstor volume: missing volume capacity"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			cmd = WithData(cmd, "ip", mock.ip)
			cmd = WithData(cmd, "volname", mock.volName)
			cmd = WithData(cmd, "capacity", mock.capacity)

			c := &cstorVolumeResize{&cstorVolumeCommand{cmd}}

			result := c.Run()

			if mock.isErr {
				if mock.errMsg != result.Error().Error() {
					t.Fatalf("Test '%s' failed: expected error: %q actual error: %q", name, mock.errMsg, result.Error().Error())
				}
			}
		})
	}
}

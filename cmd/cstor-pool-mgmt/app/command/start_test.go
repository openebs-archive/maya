package command

import (
	"testing"
)

func TestNewCStorPoolMgmt(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"start"},
	}

	cmd, err := NewCStorPoolMgmt()
	if err != nil {
		t.Errorf("Unable to Instatiate cstor-pool-mgmt")
	}
	cmds := cmd.Commands()

	if len(cmds) != len(cases) {
		t.Errorf("ExpectedCommands: %d ActualCommands: '%d'", len(cases), len(cmds))
	}

	for i, c := range cases {
		if c.use != cmds[i].Use {
			t.Errorf("ExpectedCommand: '%s' ActualCommand: '%s'", c.use, cmds[i].Use)
		}
	}

}

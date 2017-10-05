package command

import (
	"flag"
	"reflect"
	"sort"
	"testing"
)

func TestMeta_FlagSet(t *testing.T) {
	cases := []struct {
		Flags    FlagSetFlags
		Expected []string
	}{
		{
			FlagSetNone,
			[]string{},
		},
		{
			FlagSetClient,
			[]string{
				"no-color",
			},
		},
	}

	for i, tc := range cases {
		var m Meta
		fs := m.FlagSet("foo", tc.Flags)

		actual := make([]string, 0, 0)
		fs.VisitAll(func(f *flag.Flag) {
			actual = append(actual, f.Name)
		})
		sort.Strings(actual)
		sort.Strings(tc.Expected)

		if !reflect.DeepEqual(actual, tc.Expected) {
			t.Fatalf("%d: flags: %#v\n\nExpected: %#v\nGot: %#v",
				i, tc.Flags, tc.Expected, actual)
		}
	}
}

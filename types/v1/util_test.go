package v1

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseAndSubstract(t *testing.T) {
	cases := map[string]struct {
		initial, final string
		output         int64
		err            error
	}{
		"Nil string": {
			initial: "",
			final:   "",
			output:  0,
			err:     errors.New("Error in parsing string"),
		},
		"Invalid initial string": {
			initial: "1012345632342345.5",
			final:   "1222345678909889",
			output:  0,
			err:     errors.New("Error in parsing string"),
		},
		"Invalid final string": {
			initial: "1012345323455345",
			final:   "1013123454323443.5",
			output:  0,
			err:     errors.New("Error in parsing string"),
		},
		"Valid string": {
			initial: "10",
			final:   "15",
			output:  5,
			err:     nil,
		},
		"Final is less than initial": {
			initial: "15",
			final:   "10",
			output:  5,
			err:     nil,
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			_, got := ParseAndSubstract(tt.initial, tt.final)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("ParseAndSubstract(%v, %v) : expected %v, got %v", tt.initial, tt.final, tt.err, got)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	cases := map[string]struct {
		inputSlice  []string
		inputString string
		outputSlice []string
	}{
		"Remove two": {
			inputSlice:  []string{"One", "two", "three"},
			inputString: "two",
			outputSlice: []string{"One", "three"},
		},
		"Remove ERFSDFSD": {
			inputSlice:  []string{"12234234545", "342344552", "ERFSDFSD", "Wersdfw"},
			inputString: "ERFSDFSD",
			outputSlice: []string{"12234234545", "342344552", "Wersdfw"},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := Remove(tt.inputSlice, tt.inputString)
			t.Log(got, tt.outputSlice)
			if !reflect.DeepEqual(got, tt.outputSlice) {
				t.Fatalf("Remove(%v, %v) => %v, want %v", tt.inputSlice, tt.inputString, got, tt.outputSlice)
			}
		})
	}
}

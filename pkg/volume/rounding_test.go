package volume

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_RoundUpToGB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to GB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1),
		},
		{
			name:       "round k to GB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1),
		},
		{
			name:       "round Mi to GB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(2),
		},
		{
			name:       "round M to GB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(1),
		},
		{
			name:       "round G to GB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(1000),
		},
		{
			name:       "round Gi to GB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1074),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToGB(test.resource)
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

func Test_RoundUpToGiB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to GiB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1),
		},
		{
			name:       "round k to GiB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1),
		},
		{
			name:       "round Mi to GiB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1),
		},
		{
			name:       "round M to GiB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(1),
		},
		{
			name:       "round G to GiB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(932),
		},
		{
			name:       "round Gi to GiB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1000),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToGi(test.resource)
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

func Test_RoundUpToMB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to MB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(2),
		},
		{
			name:       "round k to MB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1),
		},
		{
			name:       "round Mi to MB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1049),
		},
		{
			name:       "round M to MB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(1000),
		},
		{
			name:       "round G to MB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(1000000),
		},
		{
			name:       "round Gi to MB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1073742),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToMB(test.resource)
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

func Test_RoundUpToMiB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to MiB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1),
		},
		{
			name:       "round k to MiB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1),
		},
		{
			name:       "round Mi to MiB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1000),
		},
		{
			name:       "round M to MiB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(954),
		},
		{
			name:       "round G to MiB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(953675),
		},
		{
			name:       "round Gi to MiB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1024000),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToMi(test.resource)
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

func Test_RoundUpToKB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to KB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1024),
		},
		{
			name:       "round k to KB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1000),
		},
		{
			name:       "round Mi to KB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1048576),
		},
		{
			name:       "round M to KB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(1000000),
		},
		{
			name:       "round G to KB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(1000000000),
		},
		{
			name:       "round Gi to KB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1073741824),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToKB(test.resource)
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

func Test_RoundUpToKiB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to KiB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1000),
		},
		{
			name:       "round k to KiB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(977),
		},
		{
			name:       "round Mi to KiB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1024000),
		},
		{
			name:       "round M to KiB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(976563),
		},
		{
			name:       "round G to KiB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(976562500),
		},
		{
			name:       "round Gi to KiB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1048576000),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToKi(test.resource)
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

func TestRoundUpStringToGi(t *testing.T) {
	testcases := map[string]struct {
		size           int64
		unit           string
		expectedOutput int64
		expectedError  error
	}{
		"Zi to Gi": {
			size:           int64(5),
			unit:           "Zi",
			expectedOutput: int64(5497558138880),
			expectedError:  nil,
		},
		"Z to Gi": {
			size:           int64(6),
			unit:           "Z",
			expectedOutput: int64(5592000000000),
			expectedError:  nil,
		},
		"Ei to Gi": {
			size:           int64(3),
			unit:           "Ei",
			expectedOutput: int64(3221225472),
			expectedError:  nil,
		},
		"E to Gi": {
			size:           int64(2),
			unit:           "E",
			expectedOutput: int64(1864000000),
			expectedError:  nil,
		},
		"Pi to Gi": {
			size:           int64(6),
			unit:           "Pi",
			expectedOutput: int64(6291456),
			expectedError:  nil,
		},
		"P to Gi": {
			size:           int64(5),
			unit:           "P",
			expectedOutput: int64(4660000),
			expectedError:  nil,
		},
		"Ti to Gi": {
			size:           int64(2),
			unit:           "Ti",
			expectedOutput: int64(2048),
			expectedError:  nil,
		},
		"T to Gi": {
			size:           int64(2),
			unit:           "T",
			expectedOutput: int64(1864),
			expectedError:  nil,
		},
		"G to Gi": {
			size:           int64(4),
			unit:           "G",
			expectedOutput: int64(4),
			expectedError:  nil,
		},
		"Mi to Gi": {
			size:           int64(2048),
			unit:           "Mi",
			expectedOutput: int64(2),
			expectedError:  nil,
		},
		"M to Gi": {
			size:           int64(3124),
			unit:           "M",
			expectedOutput: int64(3),
			expectedError:  nil,
		},
		"Ki to Gi": {
			size:           int64(123452),
			unit:           "Ki",
			expectedOutput: int64(0),
			expectedError:  nil,
		},
		"K to Gi": {
			size:           int64(2048),
			unit:           "K",
			expectedOutput: int64(0),
			expectedError:  nil,
		},
		"B to Gi": {
			size:           int64(2048),
			unit:           "B",
			expectedOutput: int64(0),
			expectedError:  nil,
		},
		"Invalid unit": {
			size:           int64(1233),
			unit:           "Hi",
			expectedOutput: int64(-1),
			expectedError:  fmt.Errorf("Invalid unit"),
		},
	}

	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {
			val, err := RoundUpStringToGi(test.size, test.unit)
			if val != test.expectedOutput && err == nil {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.expectedOutput)
				t.Error("unexpected rounded value")
			}
			if !reflect.DeepEqual(err, test.expectedError) {
				t.Errorf("Test case Name: %s Expected error: %v but got: %v", name, test.expectedError, err)
			}
		})
	}
}

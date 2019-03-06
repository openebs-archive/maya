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

func Test_RoundUpToBytes(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to Bytes",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1024000),
		},
		{
			name:       "round k to Bytes",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1000000),
		},
		{
			name:       "round Mi to Bytes",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1048576000),
		},
		{
			name:       "round M to Bytes",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(1000000000),
		},
		{
			name:       "round G to Bytes",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(1000000000000),
		},
		{
			name:       "round Gi to Bytes",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1073741824000),
		},
		{
			name:       "round T to Bytes",
			resource:   resource.MustParse("989T"),
			roundedVal: int64(989000000000000),
		},
		{
			name:       "round Ti to Bytes",
			resource:   resource.MustParse("1024Ti"),
			roundedVal: int64(1125899906842624),
		},
		{
			name:       "round Ei to Bytes",
			resource:   resource.MustParse("979Ei"),
			roundedVal: int64(9223372036854775807),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpToBytes(test.resource)
			if val != test.roundedVal {
				t.Logf("actual value: %d", val)
				t.Logf("expected value: %d", test.roundedVal)
				t.Error("unexpected value")
			}
		})
	}
}

func TestRoundUpStringToBytes(t *testing.T) {
	testcases := map[string]struct {
		size           uint64
		unit           string
		expectedOutput uint64
		expectedError  error
	}{
		"Ei to Bytes": {
			size:           uint64(3),
			unit:           "Ei",
			expectedOutput: uint64(3458764513820540928),
			expectedError:  nil,
		},
		"E to Bytes": {
			size:           uint64(2),
			unit:           "E",
			expectedOutput: uint64(2000000000000000000),
			expectedError:  nil,
		},
		"Pi to Bytes": {
			size:           uint64(6),
			unit:           "Pi",
			expectedOutput: uint64(6755399441055744),
			expectedError:  nil,
		},
		"P to Bytes": {
			size:           uint64(5),
			unit:           "P",
			expectedOutput: uint64(5000000000000000),
			expectedError:  nil,
		},
		"Ti to Bytes": {
			size:           uint64(2),
			unit:           "Ti",
			expectedOutput: uint64(2199023255552),
			expectedError:  nil,
		},
		"T to Bytes": {
			size:           uint64(2),
			unit:           "T",
			expectedOutput: uint64(2000000000000),
			expectedError:  nil,
		},
		"Gi to Bytes": {
			size:           uint64(7),
			unit:           "Gi",
			expectedOutput: uint64(7516192768),
			expectedError:  nil,
		},
		"G to Bytes": {
			size:           uint64(4),
			unit:           "G",
			expectedOutput: uint64(4000000000),
			expectedError:  nil,
		},
		"Mi to Bytes": {
			size:           uint64(2048),
			unit:           "Mi",
			expectedOutput: uint64(2147483648),
			expectedError:  nil,
		},
		"M to Bytes": {
			size:           uint64(3124),
			unit:           "M",
			expectedOutput: uint64(3124000000),
			expectedError:  nil,
		},
		"Ki to Bytes": {
			size:           uint64(123452),
			unit:           "Ki",
			expectedOutput: uint64(126414848),
			expectedError:  nil,
		},
		"K to Bytes": {
			size:           uint64(2048000),
			unit:           "K",
			expectedOutput: uint64(2048000000),
			expectedError:  nil,
		},
		"B to Bytes": {
			size:           uint64(2048),
			unit:           "B",
			expectedOutput: uint64(2048),
			expectedError:  nil,
		},
		"Invalid unit": {
			size:           uint64(1233),
			unit:           "Hi",
			expectedOutput: uint64(0),
			expectedError:  fmt.Errorf("Invalid unit"),
		},
	}

	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {
			val, err := RoundUpStringToBytes(test.size, test.unit)
			if val != test.expectedOutput && err == nil {
				t.Logf("actual value: %d", val)
				t.Logf("expected value: %d", test.expectedOutput)
				t.Error("unexpected rounded value")
			}
			if !reflect.DeepEqual(err, test.expectedError) {
				t.Errorf("Test case Name: %s Expected error: %v but got: %v", name, test.expectedError, err)
			}
		})
	}
}

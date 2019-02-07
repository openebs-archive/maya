package hash

import (
	"testing"
)

func TestCalculateHash(t *testing.T) {
	fakeTestList := map[string][]string{
		"list1": []string{},
		"list2": []string{"disk-1", "disk-2", "disk-3"},
		"list3": []string{"disk-4", "disk-5", "disk-6"},
	}

	fakeStruct := map[string]struct {
		poolName string
		diskList []string
	}{
		"struct1": {
			poolName: "",
			diskList: fakeTestList["list1"],
		},
		"struct2": {
			poolName: "cstor-875654-8430812-093432-4235584",
			diskList: fakeTestList["list2"],
		},
		"struct3": {
			poolName: "cstor-12321-8930812-093812-84309284",
			diskList: fakeTestList["list3"],
		},
	}
	fakeComplexStruct := map[int]struct {
		innerStructNum int
		innerStruct    struct {
			poolName string
			diskList []string
		}
	}{
		1: {
			innerStructNum: 1,
			innerStruct:    fakeStruct["struct1"],
		},
		2: {
			innerStructNum: 2,
			innerStruct:    fakeStruct["struct2"],
		},
	}
	for _, test := range fakeTestList {
		fakeHash, err := CalculateHash(test)
		if err != nil {
			t.Errorf("Failed to calculate the hash expected string but got: '%s' Error: '%v'", fakeHash, err)
		}
	}
	for _, test := range fakeStruct {
		fakeHash, err := CalculateHash(test)
		if err != nil {
			t.Errorf("Failed to calculate the hash expected string but got: '%s' Error: '%v'", fakeHash, err)
		}
	}
	for _, test := range fakeComplexStruct {
		fakeHash, err := CalculateHash(test)
		if err != nil {
			t.Errorf("Failed to calculate the hash expected string but got: '%s' Error: '%v'", fakeHash, err)
		}
	}
	fakeHash, err := CalculateHash(fakeTestList)
	if err != nil {
		t.Errorf("Failed to calculate the hash expected string but got: '%s' Error: '%v'", fakeHash, err)
	}
	fakeHash, err = CalculateHash(fakeStruct)
	if err != nil {
		t.Errorf("Failed to calculate the hash expected string but got: '%s' Error: '%v'", fakeHash, err)
	}
	fakeHash, err = CalculateHash(fakeComplexStruct)
	if err != nil {
		t.Errorf("Failed to calculate the hash expected string but got: '%s' Error: '%v'", fakeHash, err)
	}
}

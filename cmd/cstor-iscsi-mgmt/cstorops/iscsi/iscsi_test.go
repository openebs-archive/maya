package iscsi

import (
	"fmt"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// TestCheckValidIscsi tests iscsi related operations
func TestCheckValidIscsi(t *testing.T) {
	testIscsiResource := map[string]struct {
		expectedIscsiName string
		expectedError    error
		test             *apis.CStorIscsi
	}{
		"Valid-img1IscsiResource": {
			expectedIscsiName: "iscsi1",
			expectedError:    nil,
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi1",
						CacheFile: "/tmp/iscsi1.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
		"Invalid-iscsiNameEmpty": {
			expectedIscsiName: "",
			expectedError:    fmt.Errorf("Iscsiname cannot be empty"),
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "",
						CacheFile: "/tmp/iscsi1.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
		"Invalid-DiskListEmpty": {
			expectedIscsiName: "",
			expectedError:    fmt.Errorf("Disk name(s) cannot be empty"),
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{""},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi1",
						CacheFile: "/tmp/iscsi1.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
	}

	for desc, ut := range testIscsiResource {
		Obtainederr := CheckValidIscsi(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() == ut.expectedError.Error() {
				return
			}
			t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
				desc, ut.expectedError, Obtainederr)
		}

	}
}

// TestCheckValidIscsi tests iscsi related operations
func TestImportIscsiBuilder(t *testing.T) {
	testIscsiResource := map[string]struct {
		expectedCmd []string
		test        *apis.CStorIscsi
	}{
		"img1IscsiResource": {
			expectedCmd: []string{IscsiOperator + " import -c cachefile=/tmp/iscsi1.cache iscsi1"},
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi1",
						CacheFile: "/tmp/iscsi1.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},

		"img2IscsiResource": {
			expectedCmd: []string{IscsiOperator + " import -c cachefile=/tmp/iscsi2.cache iscsi2"},
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi2",
						CacheFile: "/tmp/iscsi2.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
	}

	for desc, ut := range testIscsiResource {
		obtainedCmd := importIscsiBuilder(ut.test)
		if reflect.DeepEqual(ut.expectedCmd, obtainedCmd.Args) {
			t.Fatalf("desc:%v, Commands mismatch, expected:%v, Got:%v", desc,
				ut.expectedCmd, obtainedCmd.Args[0])
		}
	}
}

// TestCheckValidIscsi tests iscsi related operations
func TestCreateIscsiBuilder(t *testing.T) {
	testIscsiResource := map[string]struct {
		expectedCmd []string
		test        *apis.CStorIscsi
	}{
		"img1IscsiResource": {
			expectedCmd: []string{IscsiOperator + " create -f -o cachefile=/tmp/iscsi1.cache iscsi1"},
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi1",
						CacheFile: "/tmp/iscsi1.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},

		"img2IscsiResource": {
			expectedCmd: []string{IscsiOperator + " create -f -o cachefile=/tmp/iscsi2.cache iscsi2"},
			test: &apis.CStorIscsi{
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi2",
						CacheFile: "/tmp/iscsi2.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
	}

	for desc, ut := range testIscsiResource {
		obtainedCmd := createIscsiBuilder(ut.test)
		if reflect.DeepEqual(ut.expectedCmd, obtainedCmd.Args) {
			t.Fatalf("desc:%v, Commands mismatch, expected:%v, Got:%v", desc,
				ut.expectedCmd, obtainedCmd.Args[0])
		}
	}
}

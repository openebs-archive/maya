package block_test

import (
	"testing"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/block"
	"github.com/openebs/maya/cmd/maya-nodebot/types/v1"
)

//SampleOutput contains only necessary fields on block disk to validate
type SampleOutput struct {
	Name       string
	Mountpoint string
	Type       string
}

//TestListBlockDevice is to check block disks with right samples
func TestListBlockDevice(t *testing.T) {
	var resJsonDecoded v1.BlockDeviceInfo
	sampleParentOutput := SampleOutput{"sda", "", "disk"}
	sampleChildrenOutput := SampleOutput{"sda1", "", "part"}

	err := block.ListBlockExec(&resJsonDecoded)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	flag := ValidateOutput(&resJsonDecoded, &sampleParentOutput, &sampleChildrenOutput)
	if flag != true {
		t.Fatalf("Invalid Output")
	}

}

//TestListBlockDevice_Negative is to check block disks with wrong samples
func TestListBlockDevice_Negative(t *testing.T) {
	var resJsonDecoded v1.BlockDeviceInfo
	sampleParentOutput := SampleOutput{"abc", "/abc", "loop"}
	sampleChildrenOutput := SampleOutput{"xyz", "/xyz", "loop"}

	err := block.ListBlockExec(&resJsonDecoded)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	flag := ValidateOutput(&resJsonDecoded, &sampleParentOutput, &sampleChildrenOutput)
	if flag == true {
		t.Fatalf("Invalid Output")
	}

}

//ValidateOutput is common function to validate samples
func ValidateOutput(resJsonDecoded *v1.BlockDeviceInfo, sampleParentOutput *SampleOutput,
	sampleChildrenOutput *SampleOutput) bool {
	for _, v := range resJsonDecoded.Blockdevices {
		if v.Type == sampleParentOutput.Type &&
			v.Mountpoint == sampleParentOutput.Mountpoint &&
			v.Name[:1] == sampleParentOutput.Name[:1] {
			return true
		}
		if v.Children != nil {
			for _, u := range v.Children {
				if u.Type == sampleChildrenOutput.Type &&
					u.Name[:1] == sampleChildrenOutput.Name[:1] {
					return true
				}
			}
		}
	}
	return false
}

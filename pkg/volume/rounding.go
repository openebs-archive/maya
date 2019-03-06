package volume

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// EB - ExaByte size
	EB = 1000 * 1000 * 1000 * 1000 * 1000 * 1000
	// Ei - ExbiByte size
	Ei = 1024 * 1024 * 1024 * 1024 * 1024 * 1024

	// PB - PetaByte size
	PB = 1000 * 1000 * 1000 * 1000 * 1000
	// Pi - PebiByte size
	Pi = 1024 * 1024 * 1024 * 1024 * 1024

	// TB - TeraByte size
	TB = 1000 * 1000 * 1000 * 1000
	// Ti - TebiByte size
	Ti = 1024 * 1024 * 1024 * 1024

	// GB - GigaByte size
	GB = 1000 * 1000 * 1000
	// Gi - GibiByte size
	Gi = 1024 * 1024 * 1024

	// MB - MegaByte size
	MB = 1000 * 1000
	// Mi - MebiByte size
	Mi = 1024 * 1024

	// KB - KiloByte size
	KB = 1000
	// Ki - KibiByte size
	Ki = 1024
)

// RoundUpToEB rounds up given quantity to chunks of EB
func RoundUpToEB(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, EB)
}

// RoundUpToEi rounds up given quantity upto chunks of Ei
func RoundUpToEi(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, Ei)
}

// RoundUpToPB rounds up given quantity to chunks of PB
func RoundUpToPB(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, PB)
}

// RoundUpToPi rounds up given quantity upto chunks of Pi
func RoundUpToPi(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, Pi)
}

// RoundUpToTB rounds up given quantity to chunks of TB
func RoundUpToTB(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, TB)
}

// RoundUpToTi rounds up given quantity upto chunks of Ti
func RoundUpToTi(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, Ti)
}

// RoundUpToGB rounds up given quantity to chunks of GB
func RoundUpToGB(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, GB)
}

// RoundUpToGi rounds up given quantity upto chunks of Gi
func RoundUpToGi(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, Gi)
}

// RoundUpToMB rounds up given quantity to chunks of MB
func RoundUpToMB(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, MB)
}

// RoundUpToMi rounds up given quantity upto chunks of Mi
func RoundUpToMi(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, Mi)
}

// RoundUpToKB rounds up given quantity to chunks of KB
func RoundUpToKB(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, KB)
}

// RoundUpToKi rounds up given quantity upto chunks of Ki
func RoundUpToKi(size resource.Quantity) int64 {
	requestBytes := size.Value()
	return roundUpSize(requestBytes, Ki)
}

// RoundUpToBytes rounds up given quantity upto chunks of Bytes
func RoundUpToBytes(size resource.Quantity) int64 {
	return size.Value()
}

// RoundUpToGBInt rounds up given quantity to chunks of GB. It returns an
// int instead of an int64 and an error if there's overflow
func RoundUpToGBInt(size resource.Quantity) (int, error) {
	requestBytes := size.Value()
	return roundUpSizeInt(requestBytes, GB)
}

// RoundUpToGiInt rounds up given quantity upto chunks of Gi. It returns an
// int instead of an int64 and an error if there's overflow
func RoundUpToGiInt(size resource.Quantity) (int, error) {
	requestBytes := size.Value()
	return roundUpSizeInt(requestBytes, Gi)
}

// RoundUpToMBInt rounds up given quantity to chunks of MB. It returns an
// int instead of an int64 and an error if there's overflow
func RoundUpToMBInt(size resource.Quantity) (int, error) {
	requestBytes := size.Value()
	return roundUpSizeInt(requestBytes, MB)
}

// RoundUpToMiBInt rounds up given quantity upto chunks of Mi. It returns an
// int instead of an int64 and an error if there's overflow
func RoundUpToMiBInt(size resource.Quantity) (int, error) {
	requestBytes := size.Value()
	return roundUpSizeInt(requestBytes, Mi)
}

// RoundUpToKBInt rounds up given quantity to chunks of KB. It returns an
// int instead of an int64 and an error if there's overflow
func RoundUpToKBInt(size resource.Quantity) (int, error) {
	requestBytes := size.Value()
	return roundUpSizeInt(requestBytes, KB)
}

// RoundUpToKiInt rounds up given quantity upto chunks of Ki. It returns an
// int instead of an int64 and an error if there's overflow
func RoundUpToKiInt(size resource.Quantity) (int, error) {
	requestBytes := size.Value()
	return roundUpSizeInt(requestBytes, Ki)
}

// roundUpSizeInt calculates how many allocation units are needed to accommodate
// a volume of given size. It returns an int instead of an int64 and an error if
// there's overflow
func roundUpSizeInt(volumeSizeBytes, allocationUnitBytes int64) (int, error) {
	roundedUp := roundUpSize(volumeSizeBytes, allocationUnitBytes)
	roundedUpInt := int(roundedUp)
	if int64(roundedUpInt) != roundedUp {
		return 0, fmt.Errorf("capacity %v is too great, casting results in integer overflow", roundedUp)
	}
	return roundedUpInt, nil
}

// roundUpSize calculates how many allocation units are needed to accommodate
// a volume of given size. E.g. when user wants 1500Mi volume, while AWS EBS
// allocates volumes in gibibyte-sized chunks,
// RoundUpSize(1500 * 1024*1024, 1024*1024*1024) returns '2'
// (2 Gi is the smallest allocatable volume that can hold 1500Mi)
func roundUpSize(volumeSizeBytes, allocationUnitBytes int64) int64 {
	roundedUp := volumeSizeBytes / allocationUnitBytes
	if volumeSizeBytes%allocationUnitBytes > 0 {
		roundedUp++
	}
	return roundedUp
}

// RoundUpStringToBytes converts the given size into bytes
func RoundUpStringToBytes(sizeVal uint64, unit string) (uint64, error) {
	var value uint64
	switch unit {
	case "Ei":
		value = sizeVal * Ei
	case "E":
		value = sizeVal * EB
	case "Pi":
		value = sizeVal * Pi
	case "P":
		value = sizeVal * PB
	case "Ti":
		value = sizeVal * Ti
	case "T":
		value = sizeVal * TB
	case "G":
		value = sizeVal * GB
	case "Gi":
		value = sizeVal * Gi
	case "Mi":
		value = sizeVal * Mi
	case "M":
		value = sizeVal * MB
	case "K":
		value = sizeVal * KB
	case "Ki":
		value = sizeVal * Ki
	case "B":
		value = sizeVal
	default:
		return uint64(0), fmt.Errorf("Invalid unit")
	}
	return value, nil
}

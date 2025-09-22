package lightbits

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

var errIntOutOfRange = errors.New("integer out of range")

// BytesToGiBString converts a uint64 number of bytes to a string in the format "XGiB".
func bytesToGiBString(bytes uint64) string {
	const bytesPerGiB = 1024 * 1024 * 1024 // 1 GiB = 1024^3 bytes
	giB := bytes / bytesPerGiB

	return fmt.Sprintf("%dGiB", giB)
}

func stringToUint64(s string) (uint64, error) {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert '%s' to uint64: %w", s, err)
	}

	return value, nil
}

func volumeStateToIsAvail(state VolumeState) bool {
	return state == VolumeStateAvailable
}

func snapshotStateToIsAvail(state SnapshotState) bool {
	return state == SnapshotStateAvailable
}

func intToUint32Checked(i int) (uint32, error) {
	if i < 0 || i > math.MaxUint32 {
		return 0, errIntOutOfRange
	}

	return uint32(i), nil
}

package lightbits

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/rs/zerolog/log"
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

func constructACLSet(volumeACL, addNodes, removeNodes []string) []string {
	// Construct a set from existing ACL values. Add and remove elements
	// as specified in the request.
	log.Info().Msgf("Update LB ACL input: [volumeACL=%s] [addNodes=%s] [removeNodes=%s]", volumeACL, addNodes, removeNodes)
	aclSet := map[string]struct{}{
		ACLNone: {},
	}
	for _, existing := range volumeACL {
		aclSet[existing] = struct{}{}
	}
	for _, add := range addNodes {
		aclSet[add] = struct{}{}
	}
	for _, remove := range removeNodes {
		delete(aclSet, remove)
	}
	if len(aclSet) > 1 {
		delete(aclSet, ACLNone)
	}

	acl := make([]string, 0, len(aclSet))
	for nqn := range aclSet {
		acl = append(acl, nqn)
	}
	log.Info().Msgf("Updated LB ACL result: %s", acl)
	
	return acl
}

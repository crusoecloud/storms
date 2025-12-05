package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

var (
	errUnsupportedSectorSize = errors.New("unsupported sector size")
	errUnsupportedSizeUnit   = errors.New("unsupported size unit")
	errUnsupportedSizeFormat = errors.New("unsupported sector size format")
)

// ParseSize converts a size string like "10GiB", "2TiB" to bytes.
func ParseSizeString(s string) (uint64, error) {
	s = strings.TrimSpace(s)

	// Match number + unit
	re := regexp.MustCompile(`^(\d+)([GT]iB)$`)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 3 {
		return 0, errUnsupportedSizeFormat
	}

	numStr, unit := matches[1], matches[2]
	num, err := strconv.ParseUint(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint :%w", err)
	}

	switch unit {
	case "GiB":
		return num * 1024 * 1024 * 1024, nil
	case "TiB":
		return num * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, errUnsupportedSizeUnit
	}
}

func ParseSectorSizeUint(u uint) (storms.SectorSizeEnum, error) {
	switch u {
	case 512:
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512, nil
	case 4096:
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512, nil
	default:
		return 0, errUnsupportedSectorSize
	}
}

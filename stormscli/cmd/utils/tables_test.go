package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

func Test_formatBytes(t *testing.T) {
	tests := []struct {
		name   string
		input  uint64
		expect string
	}{
		{
			name:   "1 byte",
			input:  1,
			expect: "1.00 B",
		},
		{
			name:   "500 bytes",
			input:  500,
			expect: "500.00 B",
		},
		{
			name:   "128 MiB",
			input:  128 * 1024 * 1024,
			expect: "128.00 MiB",
		},
		{
			name:   "512 GiB",
			input:  (512 * 1024 * 1024 * 1024) + (512 * 1024 * 1024),
			expect: "512.50 GiB",
		},
		{
			name:   "512.75 PiB",
			input:  (512 * 1024 * 1024 * 1024 * 1024 * 1024) + (768 * 1024 * 1024 * 1024 * 1024),
			expect: "512.75 PiB",
		},
		{
			name:   "16 EiB (maximum for uint64)",
			input:  (16 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024) - 1,
			expect: "16.00 EiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := formatBytes(tt.input)
			require.Equal(t, tt.expect, actual)
		})
	}
}

func Test_formatSectorSize(t *testing.T) {
	tests := []struct {
		name   string
		input  storms.SectorSizeEnum
		expect string
	}{
		{
			name:   "4096",
			input:  storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096,
			expect: "4096 B",
		},
		{
			name:   "512",
			input:  storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
			expect: "512 B",
		},
		{
			name:   "unspecified",
			input:  storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED,
			expect: "unspecified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := formatSectorSize(tt.input)
			require.Equal(t, tt.expect, actual)
		})
	}
}

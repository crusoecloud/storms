package lightbits

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_bytesToGiBString(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{
			name:     "Exactly 1 GiB",
			input:    1024 * 1024 * 1024,
			expected: "1GiB",
		},
		{
			name:     "3 GiB",
			input:    3 * 1024 * 1024 * 1024,
			expected: "3GiB",
		},
		{
			name:     "Just under 1 GiB",
			input:    (1024 * 1024 * 1024) - 1,
			expected: "0GiB",
		},
		{
			name:     "Zero bytes",
			input:    0,
			expected: "0GiB",
		},
		{
			name:     "1 TiB = 1024 GiB",
			input:    1024 * 1024 * 1024 * 1024,
			expected: "1024GiB",
		},
		{
			name:     "2.5 TiB = 2560 GiB (rounded down)",
			input:    2*1024*1024*1024*1024 + 512*1024*1024*1024,
			expected: "2560GiB",
		},
		{
			name:     "512 MiB = 0 GiB",
			input:    512 * 1024 * 1024,
			expected: "0GiB",
		},
	}

	for _, tt := range tests {
		actual := bytesToGiBString(tt.input)
		require.Equal(t, actual, tt.expected)

	}
}

func Test_stringToUint64(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  uint64
		expectErr bool
	}{
		{
			name:      "valid uint64 string",
			input:     "123456",
			expected:  123456,
			expectErr: false,
		},
		{
			name:      "maximum uint64 value",
			input:     "18446744073709551615",
			expected:  18446744073709551615,
			expectErr: false,
		},
		{
			name:      "zero",
			input:     "0",
			expected:  0,
			expectErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "negative number",
			input:     "-1",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "non-numeric characters",
			input:     "123abc",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "overflow uint64",
			input:     "18446744073709551616", // 1 more than max
			expected:  0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := stringToUint64(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, actual, tt.expected)
		})
	}
}

func Test_volumeStateToIsAvail(t *testing.T) {
	tests := []struct {
		name     string
		input    VolumeState
		expected bool
	}{
		{
			name:     "VolumeStateUnknown",
			input:    VolumeStateUnknown,
			expected: false,
		},
		{
			name:     "VolumeStateCreating",
			input:    VolumeStateCreating,
			expected: false,
		},
		{
			name:     "VolumeStateAvailable",
			input:    VolumeStateAvailable,
			expected: true,
		},
		{
			name:     "VolumeStateDeleting",
			input:    VolumeStateDeleting,
			expected: false,
		},
		{
			name:     "VolumeStateDeleted",
			input:    VolumeStateDeleted,
			expected: false,
		},
		{
			name:     "VolumeStateFailed",
			input:    VolumeStateFailed,
			expected: false,
		},
		{
			name:     "VolumeStateUpdating",
			input:    VolumeStateUpdating,
			expected: false,
		},
		{
			name:     "VolumeStateRollback",
			input:    VolumeStateRollback,
			expected: false,
		},
		{
			name:     "VolumeStateMigrating",
			input:    VolumeStateMigrating,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := volumeStateToIsAvail(tt.input)
			require.Equal(t, actual, tt.expected)
		})
	}
}

func Test_snapshotStateToIsAvail(t *testing.T) {
	tests := []struct {
		name     string
		input    SnapshotState
		expected bool
	}{
		{
			name:     "SnapshotStateUnknown",
			input:    SnapshotStateUnknown,
			expected: false,
		},

		{
			name:     "SnapshotStateCreating",
			input:    SnapshotStateCreating,
			expected: false,
		},

		{
			name:     "SnapshotStateAvailable",
			input:    SnapshotStateAvailable,
			expected: true,
		},

		{
			name:     "SnapshotStateDeleting",
			input:    SnapshotStateDeleting,
			expected: false,
		},

		{
			name:     "SnapshotStateDeleted",
			input:    SnapshotStateDeleted,
			expected: false,
		},

		{
			name:     "SnapshotStateFailed",
			input:    SnapshotStateFailed,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := snapshotStateToIsAvail(tt.input)
			require.Equal(t, actual, tt.expected)
		})
	}
}

func Test_intToUint32Checked(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    uint32
		expectedErr bool
	}{
		{
			name:        "zero value",
			input:       0,
			expected:    0,
			expectedErr: false,
		},
		{
			name:        "positive within range",
			input:       123,
			expected:    123,
			expectedErr: false,
		},
		{
			name:        "max uint32",
			input:       int(math.MaxUint32),
			expected:    math.MaxUint32,
			expectedErr: false,
		},
		{
			name:        "negative value",
			input:       -1,
			expected:    0,
			expectedErr: true,
		},
		{
			name:        "overflow beyond uint32",
			input:       int(math.MaxUint32) + 1,
			expected:    0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		actual, err := intToUint32Checked(tt.input)
		if tt.expectedErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, actual, tt.expected)

	}
}

package loadbalancer

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Address(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "10.0.0.1:8080")
	require.NoError(t, err)
	s := &DownstreamService{
		addr: addr,
	}
	require.Equal(t, addr, s.Address())
}

func Test_ActiveCount(t *testing.T) {
	s := &DownstreamService{
		activeCount: 1,
	}
	require.Equal(t, uint64(1), s.ActiveCount())
}

func Test_ErrorCountCount(t *testing.T) {
	s := &DownstreamService{
		errorCount: 1,
	}
	require.Equal(t, uint64(1), s.ErrorCount())
}

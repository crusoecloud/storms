package loadbalancer

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadBalancer_NewLoadBalancer(t *testing.T) {
	addr, err := net.ResolveIPAddr("ip", "10.0.0.1")
	require.NoError(t, err)

	lb := NewLoadBalancer(AlgorithmRoundRobin, addr)
	require.Equal(t, 1, len(lb.services))
	require.Equal(t, uint64(0), lb.dialCount)
}

func Test_LoadBalancer_Dial(t *testing.T) {
	dialFunc := func(network, address string) (net.Conn, error) {
		require.Equal(t, "tcp", network)
		require.Equal(t, "10.0.0.1:8080", address)
		return &mockConn{
			mockClose: func() error {
				return nil
			},
		}, nil
	}

	addr, err := net.ResolveTCPAddr("tcp", "10.0.0.1:8080")
	require.NoError(t, err)

	lb := LoadBalancer{
		algorithm: AlgorithmRoundRobin,
		services: []*DownstreamService{
			newDownstreamService(addr, dialFunc),
		},
		dialCount: 0,
	}
	conn, err := lb.Dial()
	require.NoError(t, err)
	require.Equal(t, uint64(1), lb.DialCount())
	require.Equal(t, uint64(1), lb.services[0].ActiveCount())
	require.Equal(t, uint64(0), lb.services[0].ErrorCount())
	err = conn.Close()
	require.NoError(t, err)
	require.Equal(t, uint64(0), lb.services[0].ActiveCount())
	require.Equal(t, uint64(0), lb.services[0].ErrorCount())
}

func Test_LoadBalancer_Dial_Error_One(t *testing.T) {
	addr1Dialed := false
	dialFunc := func(network, address string) (net.Conn, error) {
		require.Equal(t, "tcp", network)
		switch address {
		case "10.0.0.1:8080":
			addr1Dialed = true
			return nil, errors.New("FAIL")
		case "10.0.0.2:8080":
			return &mockConn{
				mockClose: func() error {
					return nil
				},
			}, nil
		}
		return nil, errors.New("FAIL")
	}

	addr1, err := net.ResolveTCPAddr("tcp", "10.0.0.1:8080")
	require.NoError(t, err)

	addr2, err := net.ResolveTCPAddr("tcp", "10.0.0.2:8080")
	require.NoError(t, err)
	lb := LoadBalancer{
		algorithm: AlgorithmRoundRobin,
		services: []*DownstreamService{
			newDownstreamService(addr1, dialFunc),
			newDownstreamService(addr2, dialFunc),
		},
		dialCount: 0,
	}
	conn, err := lb.Dial()
	require.True(t, addr1Dialed)
	require.NoError(t, err)
	require.Equal(t, uint64(1), lb.DialCount())
	require.Equal(t, uint64(0), lb.services[0].ActiveCount())
	require.Equal(t, uint64(1), lb.services[0].ErrorCount())
	require.Equal(t, uint64(1), lb.services[1].ActiveCount())
	require.Equal(t, uint64(0), lb.services[1].ErrorCount())
	err = conn.Close()
	require.NoError(t, err)
	require.Equal(t, uint64(0), lb.services[0].ActiveCount())
	require.Equal(t, uint64(1), lb.services[0].ErrorCount())
	require.Equal(t, uint64(0), lb.services[1].ActiveCount())
	require.Equal(t, uint64(0), lb.services[1].ErrorCount())
}

func Test_LoadBalancer_Dial_Error_All(t *testing.T) {
	dialFunc := func(network, address string) (net.Conn, error) {
		require.Equal(t, "tcp", network)
		return nil, errors.New("FAIL")
	}

	addr1, err := net.ResolveTCPAddr("tcp", "10.0.0.1:8080")
	require.NoError(t, err)

	addr2, err := net.ResolveTCPAddr("tcp", "10.0.0.2:8080")
	require.NoError(t, err)
	lb := LoadBalancer{
		algorithm: AlgorithmRoundRobin,
		services: []*DownstreamService{
			newDownstreamService(addr1, dialFunc),
			newDownstreamService(addr2, dialFunc),
		},
		dialCount: 0,
	}
	_, err = lb.Dial()
	require.Error(t, err)
	require.Equal(t, uint64(1), lb.DialCount())
	require.Equal(t, uint64(0), lb.services[0].ActiveCount())
	require.Equal(t, uint64(1), lb.services[0].ErrorCount())
	require.Equal(t, uint64(0), lb.services[1].ActiveCount())
	require.Equal(t, uint64(1), lb.services[1].ErrorCount())
}

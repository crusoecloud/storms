package loadbalancer

import (
	"net"
	"sync/atomic"
)

type DialFunc func(network, address string) (net.Conn, error)

// The DownstreamService type owns atomic counters for a downstream service.
type DownstreamService struct {
	addr        net.Addr
	dialFunc    DialFunc
	activeCount uint64 // The number of active connections to the service. This counter must be accessed atomically.
	errorCount  uint64 // The number of failed connections to this service. This counter must be accessed atomically.
}

func newDownstreamService(addr net.Addr, dialFunc DialFunc) *DownstreamService {
	return &DownstreamService{
		addr:        addr,
		dialFunc:    dialFunc,
		activeCount: 0,
		errorCount:  0,
	}
}

func (s *DownstreamService) Dial() (*Conn, error) {
	conn, err := s.dialFunc(s.addr.Network(), s.addr.String())
	if err != nil {
		atomic.AddUint64(&s.errorCount, 1)

		return nil, err
	}

	return newConn(s, conn), nil
}

func (s *DownstreamService) Address() net.Addr {
	return s.addr
}

// Return the number of active connections to the service. This value is incremented each time
// a connection is opened, and decremented each time a connection is closed.
func (s *DownstreamService) ActiveCount() uint64 {
	return atomic.LoadUint64(&s.activeCount)
}

// Return the number of failed connections to the service. This value is monotonically increases
// over the lifetime of the service.
func (s *DownstreamService) ErrorCount() uint64 {
	return atomic.LoadUint64(&s.errorCount)
}

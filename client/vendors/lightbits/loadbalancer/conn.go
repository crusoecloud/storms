package loadbalancer

import (
	"net"
	"sync/atomic"
)

type Conn struct {
	net.Conn
	service *DownstreamService
}

func newConn(service *DownstreamService, conn net.Conn) *Conn {
	atomic.AddUint64(&service.activeCount, 1)

	return &Conn{
		Conn:    conn,
		service: service,
	}
}

func (c *Conn) Close() error {
	// AddUint64 atomically adds delta to *addr and returns the new value.
	// To subtract a signed positive constant value c from x, do AddUint64(&x, ^uint64(c-1)).
	// In particular, to decrement x, do AddUint64(&x, ^uint64(0)).
	atomic.AddUint64(&c.service.activeCount, ^uint64(0))

	//nolint:wrapcheck // No need to wrap error for embedded interface.
	return c.Conn.Close()
}

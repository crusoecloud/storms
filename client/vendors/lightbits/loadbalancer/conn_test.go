package loadbalancer

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockConn struct {
	mockRead             func(b []byte) (n int, err error)
	mockWrite            func(b []byte) (n int, err error)
	mockClose            func() error
	mockLocalAddr        func() net.Addr
	mockRemoteAddr       func() net.Addr
	mockSetDeadline      func(t time.Time) error
	mockSetReadDeadline  func(t time.Time) error
	mockSetWriteDeadline func(t time.Time) error
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	return c.mockRead(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	return c.mockWrite(b)
}

func (c *mockConn) Close() error {
	return c.mockClose()
}

func (c *mockConn) LocalAddr() net.Addr {
	return c.mockLocalAddr()
}

func (c *mockConn) RemoteAddr() net.Addr {
	return c.mockRemoteAddr()
}

func (c *mockConn) SetDeadline(t time.Time) error {
	return c.mockSetDeadline(t)
}

func (c *mockConn) SetReadDeadline(t time.Time) error {
	return c.mockSetReadDeadline(t)
}

func (c *mockConn) SetWriteDeadline(t time.Time) error {
	return c.mockSetWriteDeadline(t)
}

func Test_Conn_Close(t *testing.T) {
	downstreamService := DownstreamService{
		activeCount: 0,
	}
	embeddedConnClosed := false
	conn := newConn(&downstreamService, &mockConn{
		mockClose: func() error {
			embeddedConnClosed = true
			return nil
		},
	})
	require.Equal(t, uint64(1), downstreamService.activeCount)
	conn.Close()
	require.Equal(t, uint64(0), downstreamService.activeCount)
	require.True(t, embeddedConnClosed)
}

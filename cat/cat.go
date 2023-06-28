package cat

import (
	"io"
	"net"
	"time"
)

type RWCNetConn struct {
	io.ReadWriteCloser
}

func (c RWCNetConn) LocalAddr() net.Addr {
	return &net.UnixAddr{}
}
func (c RWCNetConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{}
}

func (c RWCNetConn) SetDeadline(t time.Time) error {
	return nil
}
func (c RWCNetConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (c RWCNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

package j

import (
	"io"
	"net"
	"time"
)

type J struct {
	io.ReadWriteCloser
}

func (j J) LocalAddr() net.Addr {
	var null net.Addr
	return null
}
func (j J) RemoteAddr() net.Addr {
	var null net.Addr
	return null
}
func (j J) SetDeadline(t time.Time) error {
	return nil
}
func (j J) SetReadDeadline(t time.Time) error {
	return nil
}
func (j J) SetWriteDeadline(t time.Time) error {
	return nil
}

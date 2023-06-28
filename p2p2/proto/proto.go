package proto

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"portal.pc/mangi"
	"portal.pc/p2p2/dial"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

func ProtRead(x net.Conn, a multiaddr.Multiaddr, b []string) (bool, error) {
	err := binary.Write(x, binary.BigEndian, a)
	if err != nil {
		return false, err
	}
	err = binary.Write(x, binary.BigEndian, b)
	if err != nil {
		return false, err
	}
	var y [1]byte
	_, err = x.Read(y[:])
	if err != nil {
		return false, err
	}
	return y[0] == 0, nil
}
func ProtWrite(x net.Conn, f func(multiaddr.Multiaddr, []string) (func(net.Conn) error, bool, error)) error {
	var m multiaddr.Multiaddr
	err := binary.Read(x, binary.BigEndian, &m)
	if err != nil {
		return err
	}
	var n []string
	err = binary.Read(x, binary.BigEndian, &n)
	if err != nil {
		return err
	}
	cf, ok, err := f(m, n)
	if err != nil {
		return err
	}
	var y [1]byte
	if !ok {
		y[0] = 1
	}
	_, err = x.Write(y[:])
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return cf(x)
}

type P2P2 struct {
	Conns []func() (net.Conn, error)
	Me    string
}

func ProtProxy(x net.Conn, n P2P2, defaul func(net.Conn) error) error {
	return ProtWrite(x, func(m multiaddr.Multiaddr, s []string) (func(net.Conn) error, bool, error) {
		y, err := m.ValueForProtocol(Proto.Code)
		if err != nil {
			return nil, false, err
		}
		if n.Me == y {
			return defaul, true, nil
		}
		if slices.Contains(s, n.Me) {
			return nil, false, nil
		}
		s = append(s, n.Me)
		var c net.Conn
		var g errgroup.Group
		for _, v := range n.Conns {
			v := v
			g.Go(func() error {
				w, err := v()
				if err != nil {
					return err
				}
				ok, err := ProtRead(w, m, s)
				if err != nil {
					return err
				}
				if ok {
					c = w
				}
				return nil
			})
		}
		err = g.Wait()
		if err != nil {
			return nil, false, err
		}
		if c == nil {
			return nil, false, nil
		}
		return func(d net.Conn) error {
			var g errgroup.Group
			g.Go(func() error {
				_, err := io.Copy(c, d)
				return err
			})
			g.Go(func() error {
				_, err := io.Copy(d, c)
				return err
			})
			return g.Wait()
		}, true, nil
	})
}

var Proto multiaddr.Protocol = multiaddr.Protocol{
	Name: "p2p2",
	Code: 1098,
	Size: -1,
	Transcoder: multiaddr.NewTranscoderFromFunctions(func(s string) ([]byte, error) {
		return []byte(mangi.Mangle(s, '/')), nil
	}, func(b []byte) (string, error) {
		return mangi.Demangle(string(b), '/'), nil
	}, func(b []byte) error {
		return nil
	}),
}

type P2P2Key struct{}

func Init() {
	multiaddr.AddProtocol(Proto)
	dial.Bases["p2p2"] = func(ctx context.Context, a multiaddr.Multiaddr) (net.Conn, error) {
		n := ctx.Value(P2P2Key{}).(P2P2)
		var c net.Conn
		var g errgroup.Group
		for _, v := range n.Conns {
			v := v
			g.Go(func() error {
				w, err := v()
				if err != nil {
					return err
				}
				ok, err := ProtRead(w, a, []string{})
				if err != nil {
					return err
				}
				if ok {
					c = w
				}
				return nil
			})
		}
		err := g.Wait()
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, fmt.Errorf("no connection to %v", a)
		}
		return c, nil
	}
}

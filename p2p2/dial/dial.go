package dial

import (
	"context"
	"fmt"
	"net"
	"strings"

	"portal.pc/p2p2/j"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/multiformats/go-multiaddr"
)

var Bases map[string]func(ctx context.Context, a multiaddr.Multiaddr) (net.Conn, error)

var Resolvers map[string]func(ctx context.Context, c *multiaddr.Component, b net.Conn) (net.Conn, error)

func Dial(ctx context.Context, a multiaddr.Multiaddr, base net.Conn) (net.Conn, error) {
	p := []string{}
	for _, r := range a.Protocols() {
		p = append(p, r.Name)
	}
	r := strings.Join(p, ",")
	if x, ok := Bases[r]; ok && base == nil {
		return x(ctx, a)
	}
	if r == "" {
		return base, nil
	}
	t, h := multiaddr.SplitLast(a)
	x, err := Dial(ctx, t, base)
	if err != nil {
		return nil, err
	}
	s, ok := Resolvers[h.String()]
	if !ok {
		return nil, fmt.Errorf("resolver not found")
	}
	return s(ctx, h, x)
}

type LibP2P struct{}

func init() {
	Bases["ip4,tcp"] = func(ctx context.Context, m multiaddr.Multiaddr) (net.Conn, error) {
		i, err := m.ValueForProtocol(multiaddr.P_IP4)
		if err != nil {
			return nil, err
		}
		t, err := m.ValueForProtocol(multiaddr.P_TCP)
		if err != nil {
			return nil, err
		}
		return net.Dial("tcp", fmt.Sprintf("%s:%s", i, t))
	}
	Bases["ip6,tcp"] = func(ctx context.Context, m multiaddr.Multiaddr) (net.Conn, error) {
		i, err := m.ValueForProtocol(multiaddr.P_IP6)
		if err != nil {
			return nil, err
		}
		t, err := m.ValueForProtocol(multiaddr.P_TCP)
		if err != nil {
			return nil, err
		}
		return net.Dial("tcp", fmt.Sprintf("%s:%s", i, t))
	}
	Resolvers["tunnel"] = func(ctx context.Context, c *multiaddr.Component, b net.Conn) (net.Conn, error) {
		x := c.Value()
		y, err := multiaddr.NewMultiaddr(x)
		if err != nil {
			return nil, err
		}
		return Dial(ctx, y, b)
	}
	Bases["p2p"] = func(ctx context.Context, a multiaddr.Multiaddr) (net.Conn, error) {
		n := ctx.Value(LibP2P{}).(host.Host)
		i, err := a.ValueForProtocol(multiaddr.P_P2P)
		if err != nil {
			return nil, err
		}
		c, err := n.NewStream(ctx, peer.ID(i), "p2p2/p")
		return j.J{c}, err
	}
}

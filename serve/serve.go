package serve

import (
	"context"
	"encoding/hex"
	"net"
	"net/http"
	"strings"

	"portal.pc/cat"
	"portal.pc/hr2"
	"portal.pc/proxy"
	"portal.pc/rev"
	"github.com/go-chi/chi/v5"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type ServeState struct {
	Base  string
	Host  host.Host
	Canon peer.ID
}

func New(s ServeState) http.Handler {
	r := chi.NewMux()
	g := chi.NewMux()
	hr := hr2.Routes(func(r string) (chi.Router, bool) {
		if !strings.HasSuffix(r, s.Base) {
			return nil, false
		}
		if r == s.Base {
			return g, true
		}
		rr := strings.TrimSuffix(r, "."+s.Base)
		rs := strings.Split(rr, ".")
		rev.Rev(rs)
		h, err := hex.DecodeString(rs[0])
		if err != nil {
			return nil, false
		}
		i := string(h)
		hh, err := hex.DecodeString(rs[1])
		if err != nil {
			return nil, false
		}
		ii := string(hh)
		l := chi.NewMux()
		l.Mount("/", &proxy.Proxy{
			Client: &http.Client{
				Transport: &http.Transport{
					DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
						x, err := s.Host.NewStream(ctx, peer.ID(i), protocol.ID(ii))
						return cat.RWCNetConn{ReadWriteCloser: x}, err
					},
				},
			},
		})
		return l, true
	})
	r.Mount("/", hr)
	return r
}

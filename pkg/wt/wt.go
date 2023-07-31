package wt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	gostream "github.com/libp2p/go-libp2p-gostream"
	peer "github.com/libp2p/go-libp2p/core/peer"
	protocol "github.com/libp2p/go-libp2p/core/protocol"

	host "github.com/libp2p/go-libp2p/core/host"
	"github.com/portal-co/api/pkg/httpfx"
	"go.uber.org/fx"
)

type Data struct {
	// *pt.ServerInfo
	Serve host.Host
}

func Mod(name string) fx.Option {
	return fx.Options(
		fx.Provide(func(c host.Host) Data {
			return Data{c}
		}),
		fx.Provide(func(d Data) httpfx.Route {
			return httpfx.RouteFunc(name, Handle(d))
		}),
	)
}

// proxy copies data bidirectionally from one connection to another.
func proxy(local net.Conn, r io.Reader, conn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if _, err := io.Copy(conn, local); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			log.Printf("error copying ORPort to WebSocket %v", err)
		}
		local.(interface{ CloseRead() error }).CloseRead()
		conn.Close()
		wg.Done()
	}()
	go func() {
		if _, err := io.Copy(local, r); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			log.Printf("error copying WebSocket to ORPort %v", err)
		}
		local.(interface{ CloseWrite() error }).CloseWrite()
		conn.Close()
		wg.Done()
	}()

	wg.Wait()
}

const ptMethodName = "webtunnel"

// handleConn bidirectionally connects a client webtunnel connection with an ORPort.
func handleConn(ctx context.Context, conn net.Conn, r io.Reader, d Data, u *url.URL) error {
	// addr := conn.RemoteAddr().String()
	or, err := gostream.Dial(ctx, d.Serve, peer.ID(u.Query().Get("p")), protocol.ID(u.Query().Get("q")))
	if err != nil {
		return fmt.Errorf("failed to connect to ORPort: %s", err)
	}
	defer or.Close()
	proxy(or, r, conn)
	return nil
}

func Handle(d Data) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		connection := strings.ToLower(req.Header.Get("Connection"))
		upgrade := strings.ToLower(req.Header.Get("Upgrade"))
		if connection != "upgrade" || upgrade != "websocket" {
			return
		}
		w.WriteHeader(101)
		w.Header().Set("Connection", "upgrade")
		w.Header().Set("Upgrade", "websocket")
		h, i, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		c := io.MultiReader(i, h)
		err = handleConn(req.Context(), h, c, d, req.URL)
		if err != nil {
			return
		}
	}
}

package httpfx

import (
	"context"
	"fmt"
	"net"
	"net/http"

	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	protocol "github.com/libp2p/go-libp2p/core/protocol"
	"go.uber.org/fx"
	"golang.org/x/build/revdial/v2"
)

type Route struct {
	http.Handler
	Pattern string
}

func NewServeMux(routes []Route) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/$", revdial.ConnHandler())
	for _, route := range routes {
		mux.Handle(route.Pattern, route)
	}
	return mux
}

func RouteFunc(p string, f http.HandlerFunc) Route {
	return Route{
		Pattern: p,
		Handler: f,
	}
}

var HTTP = fx.Provide(fx.Annotate(NewServeMux, fx.ParamTags(`group:"routes"`)))

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"routes"`),
	)
}
func NewHTTPServerLocal(a func() (net.Listener, error), lc fx.Lifecycle, h *http.ServeMux) {
	srv := &http.Server{Handler: h}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// ln, err := net.Listen("tcp", srv.Addr)
			ln, err := a()
			if err != nil {
				return err
			}
			fmt.Println("Starting HTTP server at", srv.Addr)
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
func HTTPServeOptBase(a func() (net.Listener, error)) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle, h *http.ServeMux) {
		NewHTTPServerLocal(a, lc, h)
	})
}
func HTTPServeOpt(a string) fx.Option {
	return HTTPServeOptBase(func() (net.Listener, error) {
		return net.Listen("tcp", a)
	})
}
func HTTPServeOptHost(t protocol.ID) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle, h *http.ServeMux, s host.Host) {
		a := func() (net.Listener, error) {
			return gostream.Listen(s, t)
		}
		NewHTTPServerLocal(a, lc, h)
	})
}

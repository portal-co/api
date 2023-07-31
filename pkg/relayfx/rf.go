package relayfx

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"go.uber.org/fx"
)

var RelayLP2P = fx.Invoke(func(h host.Host, l fx.Lifecycle) error {
	var r *relay.Relay
	l.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			r, err = relay.New(h)
			return err
		},
		OnStop: func(ctx context.Context) error {
			return r.Close()
		},
	})
	return nil
})

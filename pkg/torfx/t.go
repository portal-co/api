package torfx

import (
	"context"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	"go.uber.org/fx"
)

var Mod = fx.Provide(func(c fx.Lifecycle) (*pt.ServerInfo, error) {
	var p pt.ServerInfo
	c.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			p, err = pt.ServerSetup(nil)
			return err
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
	return &p, nil
})

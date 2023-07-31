package p2pfx

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/fx"
)

var Host = fx.Provide(fx.Annotate(func(opts []libp2p.Option) (host.Host, error) {
	return libp2p.New(opts...)
}, fx.ParamTags(`group:"libp2p"`)))

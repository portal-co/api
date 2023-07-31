package fxbac

import (
	"context"

	"github.com/bacalhau-project/bacalhau/pkg/ipfs"
	"github.com/bacalhau-project/bacalhau/pkg/jobstore"
	"github.com/bacalhau-project/bacalhau/pkg/jobstore/inmemory"
	"github.com/bacalhau-project/bacalhau/pkg/node"
	"github.com/bacalhau-project/bacalhau/pkg/system"
	icore "github.com/ipfs/boxo/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/fx"
)

var IpfsClientBac = fx.Provide(func(a icore.CoreAPI) ipfs.Client {
	return ipfs.NewClient(a)
})

func NodeBac(r, c bool, s string, a uint16) fx.Option {
	x := []fx.Option{fx.Provide(func(l ipfs.Client, h host.Host, j jobstore.Store) (*node.Node, error) {
		return node.NewNode(context.Background(), node.NodeConfig{IPFSClient: l, Host: h, IsRequesterNode: r, IsComputeNode: c, CleanupManager: system.NewCleanupManager(), JobStore: j, HostAddress: "0.0.0.0", APIPort: a, ComputeConfig: node.NewComputeConfigWithDefaults(), RequesterNodeConfig: node.NewRequesterConfigWithDefaults()})
	})}
	if s == "inmemory" {
		x = append(x, fx.Provide(func() jobstore.Store {
			return inmemory.NewJobStore()
		}))
	}
	if r {
		x = append(x, fx.Provide(func(n *node.Node) *node.Requester {
			return n.RequesterNode
		}))
	}
	if c {
		x = append(x, fx.Provide(func(n *node.Node) *node.Compute {
			return n.ComputeNode
		}))
	}
	return fx.Options(x...)
}

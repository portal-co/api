package fxipfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	gopath "path"

	icore "github.com/ipfs/boxo/coreiface"
	"github.com/ipfs/boxo/coreiface/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/node"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	p2phttp "github.com/libp2p/go-libp2p-http"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/portal-co/remount"
	"go.uber.org/fx"
)

// var _typeOfIn = reflect.TypeOf(fx.In{})

// func Extract2(target interface{}) fx.Option {
// 	v := reflect.ValueOf(target)

// 	if t := v.Type(); t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
// 		return fx.Error(fmt.Errorf("Extract expected a pointer to a struct, got a %v", t))
// 	}
// 	u := v.Type()

// 	v = v.Elem()
// 	t := v.Type()

// 	// We generate a function which accepts a single fx.In struct as an
// 	// argument. This struct contains all exported fields of the target
// 	// struct.

// 	// Fields of the generated fx.In struct.
// 	fields := make([]reflect.StructField, 0, t.NumField()+1)

// 	// Anonymous dig.In field.
// 	fields = append(fields, reflect.StructField{
// 		Name:      _typeOfIn.Name(),
// 		Anonymous: true,
// 		Type:      _typeOfIn,
// 	})

// 	// List of values in the target struct aligned with the fields of the
// 	// generated struct.
// 	//
// 	// So for example, if the target is,
// 	//
// 	// 	var target struct {
// 	// 		Foo io.Reader
// 	// 		bar []byte
// 	// 		Baz io.Writer
// 	// 	}
// 	//
// 	// The generated struct has the shape,
// 	//
// 	// 	struct {
// 	// 		fx.In
// 	//
// 	// 		F0 io.Reader
// 	// 		F2 io.Writer
// 	// 	}
// 	//
// 	// And `targets` is,
// 	//
// 	// 	[
// 	// 		target.Field(0),  // Foo io.Reader
// 	// 		target.Field(2),  // Baz io.Writer
// 	// 	]
// 	//
// 	// As we iterate through the fields of the generated struct, we can copy
// 	// the value into the corresponding value in the targets list.
// 	targets := make([]int, 0, t.NumField())

// 	for i := 0; i < t.NumField(); i++ {
// 		f := t.Field(i)

// 		// Skip unexported fields.
// 		if f.Anonymous {
// 			// If embedded, StructField.PkgPath is not a reliable indicator of
// 			// whether the field is exported. See
// 			// https://github.com/golang/go/issues/21122

// 			t := f.Type
// 			if t.Kind() == reflect.Ptr {
// 				t = t.Elem()
// 			}

// 			if !isExported(t.Name()) {
// 				continue
// 			}
// 		} else if f.PkgPath != "" {
// 			continue
// 		}

// 		// We don't copy over names or embedded semantics.
// 		fields = append(fields, reflect.StructField{
// 			Name: fmt.Sprintf("F%d", i),
// 			Type: f.Type,
// 			Tag:  f.Tag,
// 		})
// 		targets = append(targets, i)
// 	}

// 	// Equivalent to,
// 	//
// 	// 	func(r struct {
// 	// 		fx.In
// 	//
// 	// 		F1 Foo
// 	// 		F2 Bar
// 	// 	}) {
// 	// 		target.Foo = r.F1
// 	// 		target.Bar = r.F2
// 	// 	}

// 	fn := reflect.MakeFunc(
// 		reflect.FuncOf(
// 			[]reflect.Type{reflect.StructOf(fields)},
// 			[]reflect.Type{u}, /* results */
// 			false,             /* variadic */
// 		),
// 		func(args []reflect.Value) []reflect.Value {
// 			result := args[0]
// 			s := reflect.New(t)
// 			for i := 1; i < result.NumField(); i++ {
// 				s.Elem().Field(targets[i-1]).Set(result.Field(i))
// 			}
// 			return []reflect.Value{s}
// 		},
// 	)

// 	return fx.Provide(fn.Interface())
// }

// // isExported reports whether the identifier is exported.
// func isExported(id string) bool {
// 	r, _ := utf8.DecodeRuneInString(id)
// 	return unicode.IsUpper(r)
// }

// var Node = fx.Options(Extract2(&core.IpfsNode{}), fx.Invoke(func(l fx.Lifecycle, n *core.IpfsNode) {
// 	l.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			return n.Bootstrap(bootstrap.DefaultBootstrapConfig)
// 		},
// 		OnStop: func(ctx context.Context) error {
// 			return nil
// 		},
// 	})
// }))

// func NodeIpfs(cfg *node.BuildCfg) fx.Option {
// 	return node.IPFS(context.Background(), cfg)
// }

func Import(ctx context.Context, a icore.CoreAPI, c io.Reader) error {
	br, err := car.NewBlockReader(c)
	if err != nil {
		return err
	}
	for {
		b, err := br.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		r := bytes.NewBuffer(b.RawData())
		_, err = a.Block().Put(ctx, r)
		if err != nil {
			return err
		}
	}
}

type NodeDumper struct {
	Import func(c io.Reader) error
	Export func(root string, w io.Writer) error
}

var Dump = fx.Provide(func(i icore.CoreAPI, l linking.LinkSystem) NodeDumper {
	return NodeDumper{
		Import: func(c io.Reader) error {
			return Import(context.Background(), i, c)
		},
		Export: func(root string, w io.Writer) error {
			c, err := cid.Parse(root)
			if err != nil {
				return err
			}
			_, err = car.TraverseV1(context.Background(), &l, c, nil, w)
			return err
		},
	}
})

var Node = fx.Provide(func(l fx.Lifecycle, c *node.BuildCfg) (*core.IpfsNode, error) {
	n, err := core.NewNode(context.Background(), c)
	if err != nil {
		return nil, err
	}
	init := true
	l.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !init {
				err := n.Close()
				if err != nil {
					return err
				}
				m, err := core.NewNode(ctx, c)
				if err != nil {
					return err
				}
				*n = *m
			}
			init = false
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return n.Close()
		},
	})
	return n, nil
})

type MountPath string

var Mount = fx.Provide(func(i icore.CoreAPI, l fx.Lifecycle) (MountPath, error) {
	p := gopath.Join(os.Getenv("HOME"), "portal-ipfs")
	exec.Command("fusermount", "-u", p).Run()
	err := os.MkdirAll(p, 0777)
	if err != nil {
		return "", err
	}
	m, err := remount.Mount(remount.I{i}, p)
	if err != nil {
		return "", err
	}
	l.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if m == nil {
				n, err := remount.Mount(remount.I{i}, p)
				m = n
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			defer func() { m = nil }()
			return m()
		},
	})
	return MountPath(p), nil
})

var CoreAPI = fx.Provide(func(n *core.IpfsNode) (icore.CoreAPI, error) {
	return coreapi.NewCoreAPI(n)
})

var Cfg = fx.Provide(func() *node.BuildCfg {
	return &node.BuildCfg{}
})

func RemoteCoreAPI(x string) fx.Option {
	return fx.Provide(func(p host.Host) (icore.CoreAPI, error) {
		tr := &http.Transport{}
		tr.RegisterProtocol("libp2p", p2phttp.NewTransport(p))
		client := &http.Client{Transport: tr}
		return rpc.NewURLApiWithClient(fmt.Sprintf("libp2p://%s", x), client)
	})
}

var MerkelDAGBlock = fx.Provide(func(i icore.CoreAPI) (ipld.BlockReadOpener, ipld.BlockWriteOpener) {
	return func(lc linking.LinkContext, l datamodel.Link) (io.Reader, error) {
			theCid, ok := l.(cidlink.Link)
			if !ok {
				return nil, fmt.Errorf("attempted to load a non CID link: %v", l)
			}
			block, err := i.Block().Get(lc.Ctx, path.IpldPath(theCid.Cid))
			if err != nil {
				return nil, fmt.Errorf("error loading %v: %v", theCid.String(), err)
			}
			return block, nil
		}, func(lc linking.LinkContext) (io.Writer, linking.BlockWriteCommitter, error) {
			buf := bytes.Buffer{}
			return &buf, func(lnk ipld.Link) error {
				// _, err := shell.BlockPut(buf.Bytes(), "cbor", "sha3-384", lb.MhLength)
				_, err := i.Block().Put(lc.Ctx, &buf)
				return err
			}, nil
		}
})
var MerkelDAGLinkSystem = fx.Provide(func(r ipld.BlockReadOpener, w ipld.BlockWriteOpener) linking.LinkSystem {
	c := cidlink.DefaultLinkSystem()
	c.StorageReadOpener = r
	c.StorageWriteOpener = w
	return c
})
var MerkelDAG = fx.Options(MerkelDAGBlock, MerkelDAGLinkSystem)

// func FillDefaults(cfg *node.BuildCfg) error {
// 	if cfg.Repo != nil && cfg.NilRepo {
// 		return errors.New("cannot set a Repo and specify nilrepo at the same time")
// 	}

// 	if cfg.Repo == nil {
// 		var d ds.Datastore
// 		if cfg.NilRepo {
// 			d = ds.NewNullDatastore()
// 		} else {
// 			d = ds.NewMapDatastore()
// 		}
// 		r, err := defaultRepo(dsync.MutexWrap(d))
// 		if err != nil {
// 			return err
// 		}
// 		cfg.Repo = r
// 	}

// 	if cfg.Routing == nil {
// 		cfg.Routing = libp2p.DHTOption
// 	}

// 	if cfg.Host == nil {
// 		cfg.Host = libp2p.DefaultHostOption
// 	}

// 	return nil
// }

// // options creates fx option group from this build config
// func Options(cfg *node.BuildCfg, ctx context.Context) (fx.Option, *cfg.Config) {
// 	err := FillDefaults(cfg)
// 	if err != nil {
// 		return fx.Error(err), nil
// 	}

// 	repoOption := fx.Provide(func(lc fx.Lifecycle) repo.Repo {
// 		lc.Append(fx.Hook{
// 			OnStop: func(ctx context.Context) error {
// 				return cfg.Repo.Close()
// 			},
// 		})

// 		return cfg.Repo
// 	})

// 	metricsCtx := fx.Provide(func() helpers.MetricsCtx {
// 		return helpers.MetricsCtx(ctx)
// 	})

// 	hostOption := fx.Provide(func() libp2p.HostOption {
// 		return cfg.Host
// 	})

// 	routingOption := fx.Provide(func() libp2p.RoutingOption {
// 		return cfg.Routing
// 	})

// 	conf, err := cfg.Repo.Config()
// 	if err != nil {
// 		return fx.Error(err), nil
// 	}

// 	return fx.Options(
// 		repoOption,
// 		hostOption,
// 		routingOption,
// 		metricsCtx,
// 	), conf
// }

// func defaultRepo(dstore repo.Datastore) (repo.Repo, error) {
// 	c := cfg.Config{}
// 	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
// 	if err != nil {
// 		return nil, err
// 	}

// 	pid, err := peer.IDFromPublicKey(pub)
// 	if err != nil {
// 		return nil, err
// 	}

// 	privkeyb, err := crypto.MarshalPrivateKey(priv)
// 	if err != nil {
// 		return nil, err
// 	}

// 	c.Bootstrap = cfg.DefaultBootstrapAddresses
// 	c.Addresses.Swarm = []string{"/ip4/0.0.0.0/tcp/4001", "/ip4/0.0.0.0/udp/4001/quic"}
// 	c.Identity.PeerID = pid.Pretty()
// 	c.Identity.PrivKey = base64.StdEncoding.EncodeToString(privkeyb)

// 	return &repo.Mock{
// 		D: dstore,
// 		C: c,
// 	}, nil
// }

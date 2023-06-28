package money

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"portal.pc/orac"
	"github.com/gochain/gochain/v4/accounts/abi"
	"github.com/gochain/gochain/v4/common"
	"github.com/gochain/gochain/v4/core/types"
	"github.com/gochain/gochain/v4/rlp"
	"github.com/gochain/gochain/v4/rpc"
	"github.com/gochain/web3"
	"github.com/holiman/uint256"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

type MOracle[I any, O any] struct {
	orac.Pipe[I, O]
	rpc.Client
	ItsABI abi.ABI
	It     string
	Us     string
	OurABI abi.ABI
}

func (m MOracle[I, O]) Start(ctx context.Context, cost func(O) uint256.Int, addr func(I) common.Address) error {
	var g errgroup.Group
	n := make(map[string]bool)
	done := ctx.Done()
	g.Go(func() error {
		for {
			select {
			case <-done:
				return fmt.Errorf("Cancelled")
			case o := <-m.OutPipe:
				if o.Response {
					p, err := m.OurABI.Pack("stopCost", o.Id, cost(o.Data))
					if err != nil {
						continue
					}
					s := types.NewTransaction(0, common.HexToAddress(m.Us), nil, 0, nil, p)
					var t []byte
					err = m.Client.CallContext(ctx, &t, "eth_signTransaction", s)
					if err != nil {
						continue
					}
					err = rlp.DecodeBytes(t, &s)
					if err != nil {
						continue
					}
					err = web3.SendTransaction(ctx, web3.NewClient(&m.Client), s)
					if err != nil {
						continue
					}
				}
				jo, err := json.Marshal(o)
				if err != nil {
					continue
				}
				p, err := m.ItsABI.Pack("outerInput", jo)
				if err != nil {
					continue
				}
				s := types.NewTransaction(0, common.HexToAddress(m.It), nil, 0, nil, p)
				var t []byte
				err = m.Client.CallContext(ctx, &t, "eth_signTransaction", s)
				if err != nil {
					continue
				}
				err = rlp.DecodeBytes(t, &s)
				if err != nil {
					continue
				}
				err = web3.SendTransaction(ctx, web3.NewClient(&m.Client), s)
				if err != nil {
					continue
				}
			}
		}
	})
	g.Go(func() error {
		for {
			var o string
			p, err := web3.CallConstantFunction(ctx, web3.NewClient(&m.Client), m.ItsABI, m.It, "getDataHex")
			if err != nil {
				continue
			}
			o = p[0].(string)
			if n[o] {
				continue
			}
			n[o] = true
			b, err := hex.DecodeString(strings.TrimPrefix("0x", o))
			if err != nil {
				continue
			}
			var hs string
			err = m.CallContext(ctx, &hs, "web3_sha3", fmt.Sprintf("0x%x", b))
			if err != nil {
				continue
			}
			h, err := hex.DecodeString(strings.TrimPrefix("0x", hs))
			if err != nil {
				continue
			}
			var x struct {
				Id       uint256.Int
				Response bool
				Data     I
			}
			err = json.Unmarshal(b[3:], &x)
			if err != nil {
				continue
			}
			if !slices.Equal(b[0:2], h[0:2]) && !x.Response {
				continue // no pow
			}
			if !x.Response {
				p, err := m.OurABI.Pack("startCost", x.Id, addr(x.Data))
				if err != nil {
					continue
				}
				s := types.NewTransaction(0, common.HexToAddress(m.Us), nil, 0, nil, p)
				var t []byte
				err = m.Client.CallContext(ctx, &t, "eth_signTransaction", s)
				if err != nil {
					continue
				}
				err = rlp.DecodeBytes(t, &s)
				if err != nil {
					continue
				}
				err = web3.SendTransaction(ctx, web3.NewClient(&m.Client), s)
				if err != nil {
					continue
				}
			}
			select {
			case m.InPipe <- x:
			case <-done:
				return fmt.Errorf("Cancelled")
			}
		}
	})
	return g.Wait()
}

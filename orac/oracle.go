package orac

import (
	"context"
	"math/rand"

	"github.com/holiman/uint256"
	"golang.org/x/sync/errgroup"
)

type Oracle[I any, O any] interface {
	In(ctx context.Context) (struct {
		Id       uint256.Int
		Response bool
		Data     I
	}, error)
	Resend(context.Context, struct {
		Id       uint256.Int
		Response bool
		Data     I
	}) error
	Out(ctx context.Context, data struct {
		Id       uint256.Int
		Response bool
		Data     O
	}) error
}

type Pipe[I any, O any] struct {
	InPipe chan struct {
		Id       uint256.Int
		Response bool
		Data     I
	}
	OutPipe chan struct {
		Id       uint256.Int
		Response bool
		Data     O
	}
}

func NewPipe[I any, O any](pi *Pipe[I, O], po *Pipe[O, I]) {
	ci := make(chan struct {
		Id       uint256.Int
		Response bool
		Data     I
	})
	co := make(chan struct {
		Id       uint256.Int
		Response bool
		Data     O
	})
	*pi = Pipe[I, O]{
		InPipe:  ci,
		OutPipe: co,
	}
	*po = Pipe[O, I]{
		InPipe:  co,
		OutPipe: ci,
	}
}

func (p Pipe[I, O]) In(ctx context.Context) (struct {
	Id       uint256.Int
	Response bool
	Data     I
}, error) {
	return <-p.InPipe, nil
}

func (p Pipe[I, O]) Resend(ctx context.Context, x struct {
	Id       uint256.Int
	Response bool
	Data     I
}) error {
	p.InPipe <- x
	return nil
}

func (p Pipe[I, O]) Out(ctx context.Context, data struct {
	Id       uint256.Int
	Response bool
	Data     O
}) error {
	p.OutPipe <- data
	return nil
}

func ReadRequest[I any, O any](ctx context.Context, o Oracle[I, O]) (struct {
	Id   uint256.Int
	Data I
}, error) {
	for {
		d, err := o.In(ctx)
		if err != nil {
			return struct {
				Id   uint256.Int
				Data I
			}{}, err
		}
		if d.Response {
			defer o.Resend(ctx, d)
			continue
		}
		return struct {
			Id   uint256.Int
			Data I
		}{
			Id:   d.Id,
			Data: d.Data,
		}, nil
	}
}
func Proxy[I any, O any](ctx context.Context, a Oracle[I, O], b Oracle[O, I]) error {
	var g errgroup.Group
	g.Go(func() error {
		return Serve(ctx, a, func(i I) (O, error) {
			return SendToContract(ctx, b, i)
		})
	})
	g.Go(func() error {
		return Serve(ctx, b, func(i O) (I, error) {
			return SendToContract(ctx, a, i)
		})
	})
	return g.Wait()
}
func Serve[I any, O any](ctx context.Context, o Oracle[I, O], fun func(I) (O, error)) error {
	for {
		q, err := ReadRequest(ctx, o)
		if err != nil {
			return err
		}
		t, err := fun(q.Data)
		if err != nil {
			return err
		}
		err = o.Out(ctx, struct {
			Id       uint256.Int
			Response bool
			Data     O
		}{
			Response: true,
			Id:       q.Id,
			Data:     t,
		})
		if err != nil {
			return err
		}
	}
}
func ReadResponse[I any, O any](ctx context.Context, o Oracle[I, O], me uint256.Int) (I, error) {
	var null I
	for {
		d, err := o.In(ctx)
		if err != nil {
			return null, err
		}
		if !d.Response {
			defer o.Resend(ctx, d)
			continue
		}
		if d.Id != me {
			defer o.Resend(ctx, d)
			continue
		}
		return d.Data, nil
	}
}
func SendToContract[I any, O any](ctx context.Context, o Oracle[I, O], i O) (I, error) {
	var d uint256.Int
	var null I
	for i := 0; i < 4; i++ {
		d[i] = rand.Uint64()
	}
	err := o.Out(ctx, struct {
		Id       uint256.Int
		Response bool
		Data     O
	}{
		Data:     i,
		Id:       d,
		Response: false,
	})
	if err != nil {
		return null, err
	}
	return ReadResponse(ctx, o, d)
}

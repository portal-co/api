package cell

import "portal.pc/jade/erroror"

type Cell[T any] struct {
	cha chan T
}

func NewCell[T any](r *Cell[T]) {
	*r = Cell[T]{cha: make(chan T)}
}

func NewFell[T any](r *Fell[T]) {
	*r = Fell[T]{cha: make(chan erroror.ErrorOr[T])}
}

func (c Cell[T]) When(f func(T)) {
	go func() {
		x := <-c.cha
		f(x)
	}()
}
func (c Cell[T]) Set(val T) {
	go func() {
		for {
			c.cha <- val
		}
	}()
}
func (c Cell[T]) Val() T {
	return <-c.cha
}
func FlatMap[T any, U any](c Cell[U], v func(U) Cell[T], r Cell[T]) Cell[T] {
	c.When(func(u U) {
		v(u).When(func(t T) {
			r.Set(t)
		})
	})
	return r
}

type Fell[T any] Cell[erroror.ErrorOr[T]]

func (f Fell[T]) When(n func(T, error)) {
	Cell[erroror.ErrorOr[T]](f).When(func(eo erroror.ErrorOr[T]) {
		// if eo.Error == nil {
		n(eo.Value, eo.Error)
		// }
	})
}
func (f Fell[T]) Val() (T, error) {
	return Cell[erroror.ErrorOr[T]](f).Val().Run()
}
func (f Fell[T]) Set(v T) {
	Cell[erroror.ErrorOr[T]](f).Set(erroror.ErrorOr[T]{Value: v})
}
func (f Fell[T]) Fail(e error) {
	Cell[erroror.ErrorOr[T]](f).Set(erroror.ErrorOr[T]{Error: e})
}
func MapE[T any, U any](f Fell[T], g Fell[U], fun func(T) (U, error)) {
	f.When(func(t T, err error) {
		if err != nil {
			g.Fail(err)
			return
		}
		x, err := fun(t)
		if err != nil {
			g.Fail(err)
		} else {
			g.Set(x)
		}
	})
}
func FlatMapE[T any, U any](c Fell[U], v func(U) Fell[T], r Fell[T]) Fell[T] {
	c.When(func(u U, err error) {
		if err != nil {
			r.Fail(err)
		}
		v(u).When(func(t T, err error) {
			if err != nil {
				r.Fail(err)
			}
			r.Set(t)
		})
	})
	return r
}
func Lift[T any](c Cell[T], f Fell[T]) {
	c.When(f.Set)
}

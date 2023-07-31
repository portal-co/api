package q

type Q[T any] chan T

func Pop[T any, U any](q Q[T], fun func(T) (U, bool)) U {
	for {
		v := <-q
		u, ok := fun(v)
		if ok {
			return u
		}
		defer func() { q <- v }()
	}
}
func Get[T any, U any](q Q[T], fun func(T) (U, bool)) U {
	for {
		v := <-q
		u, ok := fun(v)
		defer func() { q <- v }()
		if ok {
			return u
		}
	}
}
func (q Q[T]) Put(val T) {
	go func() {
		q <- val
	}()
}

type Cell[T any] chan T

func (c Cell[T]) When(f func(T)) {
	go func() {
		v := <-c
		go c.Trigger(v)
		f(v)
	}()
}
func (c Cell[T]) Trigger(v T) {
	c <- v
}

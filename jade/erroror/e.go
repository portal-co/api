package erroror

type ErrorOr[T any] struct {
	Value T
	Error error
}

func New[T any](value T, err error) ErrorOr[T] {
	return ErrorOr[T]{
		Value: value,
		Error: err,
	}
}
func (t ErrorOr[T]) Run() (T, error) {
	return t.Value, t.Error
}
func FlatMap[T any, U any](t ErrorOr[T], u func(T) ErrorOr[U]) ErrorOr[U] {
	if t.Error != nil {
		var null U
		return New(null, t.Error)
	}
	return u(t.Value)
}

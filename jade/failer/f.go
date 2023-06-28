package failer

import "go.starlark.net/starlark"

func Fail(t *starlark.Thread, err error) {
	l := t.Local("failer").(func(error))
	for {
		l(err)
	}
}

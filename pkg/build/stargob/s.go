package stargob

import (
	"encoding/gob"

	"go.starlark.net/starlark"
)

func init() {
	for _, x := range []interface{}{
		starlark.None, starlark.Bool(false),
		starlark.Float(0.0), starlark.String(""), starlark.StringDict{}, starlark.Tuple{},
	} {
		gob.Register(x)
	}
}

package build

import (
	"portal.pc/jade/cell"
	"go.starlark.net/starlark"
)

func Globals() starlark.StringDict {
	d := starlark.StringDict{}
	cell.InitStell(d)
	return d
}

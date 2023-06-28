package cell

import (
	"fmt"

	"portal.pc/jade/failer"
	"go.starlark.net/starlark"
)

type Stell Fell[starlark.Value]

func (s Stell) String() string {
	return "cell"
}
func (s Stell) Type() string {
	return "cell"
}
func (s Stell) Freeze() {

}
func (s Stell) Truth() starlark.Bool {
	return true
}
func (s Stell) Hash() (uint32, error) {
	v, err := Fell[starlark.Value](s).Val()
	if err != nil {
		return 0, err
	}
	return v.Hash()
}
func (s Stell) AttrNames() []string {
	return []string{"value", "set", "when"}
}
func (s Stell) Attr(name string) (starlark.Value, error) {
	if name == "value" {
		return Fell[starlark.Value](s).Val()
	}
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		switch name {
		case "set":
			var v starlark.Value
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &v); err != nil {
				return nil, err
			}
			Fell[starlark.Value](s).Set(v)
			return s, nil
		case "when":
			var v starlark.Value
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "do", &v); err != nil {
				return nil, err
			}
			Fell[starlark.Value](s).When(func(w starlark.Value, err error) {
				if err != nil {
					failer.Fail(thread, err)
				} else {
					_, err := starlark.Call(thread, v, starlark.Tuple{w}, []starlark.Tuple{{starlark.String("value"), w}})
					if err != nil {
						failer.Fail(thread, err)
					}
				}
			})
			return s, nil
		default:
			return nil, fmt.Errorf("invalid attribute")
		}
	}), nil
}
func InitStell(d starlark.StringDict) {
	d["cell"] = starlark.NewBuiltin("cell", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var c Fell[starlark.Value]
		NewFell(&c)
		return Stell(c), nil
	})
}

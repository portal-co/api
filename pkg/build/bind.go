package build

import (
	"fmt"

	"github.com/portal-co/api/pkg/label"
	"github.com/portal-co/api/pkg/sandbox"
	"go.starlark.net/starlark"
)

type Binder interface {
	Bind(name string, target func(Binder) (Target, error), transformation func(Cfg) (Cfg, error)) error
	WithCfg(transform func(Cfg) (Cfg, error)) (Binder, error)
	Cfg() Cfg
	Name(name string) label.BuildLabel
	Resolve(l label.BuildLabel) (Resolver, string, error)
	Runner() sandbox.Runner
}
type Resolver interface {
	Get(name string, cfg func(Cfg) (Cfg, error)) (Target, error)
}
type Target struct {
	Data    starlark.Value
	Default string
}
type HashBinder struct {
	Inner Binder
	HName string
}

func (h HashBinder) Bind(name string, target func(Binder) (Target, error), transformation func(Cfg) (Cfg, error)) error {
	return h.Inner.Bind(h.HName+"#"+name, target, transformation)
}
func (h HashBinder) Name(name string) label.BuildLabel {
	return h.Inner.Name(h.HName + "#" + name)
}
func (h HashBinder) Runner() sandbox.Runner {
	return h.Inner.Runner()
}
func (h HashBinder) WithCfg(transform func(Cfg) (Cfg, error)) (Binder, error) {
	i, err := h.Inner.WithCfg(transform)
	if err != nil {
		return nil, err
	}
	return HashBinder{Inner: i, HName: h.HName}, nil
}
func (h HashBinder) Cfg() Cfg {
	return h.Inner.Cfg()
}
func (h HashBinder) Resolve(l label.BuildLabel) (Resolver, string, error) {
	return h.Inner.Resolve(l)
}

type BinderWrapper struct {
	Inner Binder
}

func (b BinderWrapper) String() string {
	return "ctx"
}
func (b BinderWrapper) Type() string {
	return "ctx"
}
func (b BinderWrapper) Freeze() {
}
func (b BinderWrapper) Truth() starlark.Bool {
	return true
}
func (b BinderWrapper) Hash() (uint32, error) {
	return 0, fmt.Errorf("not hashable")
}
func (b BinderWrapper) AttrNames() []string {
	return []string{"bind", "cfg_map", "cfg", "run", "target", "wrap"}
}
func Configurer(thread *starlark.Thread, cfg starlark.Value) func(Cfg) (Cfg, error) {
	return func(c Cfg) (Cfg, error) {
		dd := starlark.NewDict(len(c.Items()))
		for _, k := range c.Items() {
			dd.SetKey(k[0], k[1])
		}
		d, err := starlark.Call(thread, cfg, starlark.Tuple{dd}, []starlark.Tuple{})
		if err != nil {
			return Cfg{}, err
		}
		b, ok := d.(starlark.IterableMapping)
		if !ok {
			return Cfg{}, fmt.Errorf("invalid type")
		}
		dd = starlark.NewDict(len(b.Items()))
		for _, k := range b.Items() {
			dd.SetKey(k[0], k[1])
		}
		return Cfg{Dict: dd}, nil
	}
}
func (b BinderWrapper) Attr(name string) (starlark.Value, error) {
	if name == "cfg" {
		return b.Inner.Cfg().Dict, nil
	}
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		switch name {
		case "target":
			var s string
			var cfg starlark.Value
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &s, "cfg", &cfg); err != nil {
				return nil, err
			}
			l := label.ParseBuildLabel(s, thread.Name)
			r, s, err := b.Inner.Resolve(l)
			if err != nil {
				return nil, err
			}
			t, err := r.Get(s, Configurer(thread, cfg))
			if err != nil {
				return nil, err
			}
			return t.Data, nil
		case "wrap":
			var s string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &s); err != nil {
				return nil, err
			}
			return starlark.String(b.Inner.Name(s).String()), nil
		case "run":
			var deps *starlark.Dict
			var name string
			var cmd *starlark.List
			var path string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "deps", &deps, "name", &name, "cmd", &cmd, "path", &path); err != nil {
				return nil, err
			}
			s := make(map[string]string)
			for _, k := range deps.Keys() {
				d, _, _ := deps.Get(k)
				s[string(k.(starlark.String))] = string(d.(starlark.String))
			}
			cmd2 := []string{}
			for i := 0; i < cmd.Len(); i++ {
				v := cmd.Index(i)
				cmd2 = append(cmd2, string(v.(starlark.String)))
			}
			a, err := b.Inner.Runner().Run(s, cmd2, []string{"."})
			if err != nil {
				return nil, err
			}
			return starlark.String(a), nil
		case "bind":
			var name string
			var binder starlark.Callable
			var cfg starlark.Value = starlark.None
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "binder", &binder, "transition??", &cfg); err != nil {
				return nil, err
			}
			var f func(Cfg) (Cfg, error)
			if cfg != starlark.None {
				f = Configurer(thread, cfg)
			}
			return starlark.None, b.Inner.Bind(name, func(b Binder) (Target, error) {
				x, err := starlark.Call(thread, binder, starlark.Tuple{BinderWrapper{Inner: b}}, []starlark.Tuple{})
				if err != nil {
					return Target{}, err
				}
				var s starlark.Value
				a, ok := x.(starlark.HasAttrs)
				if !ok {
					b, ok := x.(starlark.Mapping)
					if !ok {
						return Target{}, fmt.Errorf("invalid type")
					}
					s, ok, err = b.Get(starlark.String("default"))
					if err != nil {
						return Target{}, err
					}
					if !ok {
						return Target{}, fmt.Errorf("invalid type")
					}
				} else {
					s, err = a.Attr("default")
					if err != nil {
						return Target{}, err
					}
				}
				t, ok := s.(starlark.String)
				if !ok {
					return Target{}, fmt.Errorf("invalid type")
				}
				return Target{Data: x, Default: string(t)}, nil
			}, f)
		case "cfg_map":
			var fun starlark.Callable
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "fun", &fun); err != nil {
				return nil, err
			}
			c, err := b.Inner.WithCfg(Configurer(thread, fun))
			return BinderWrapper{Inner: c}, err
		default:
			return nil, fmt.Errorf("invalid context attribute")
		}
	}), nil
}

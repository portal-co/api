package build

import (
	"crypto/sha256"
	"strings"

	"github.com/portal-co/api/pkg/label"
	"github.com/portal-co/api/pkg/q"
	"github.com/portal-co/api/pkg/sandbox"
)

type PackageBinder struct {
	Lab      label.BuildLabel
	F        Cfg
	R        sandbox.Runner
	ResolveF func(l label.BuildLabel) (Resolver, string, error)
	q.Q[PackageBind]
}

func (p PackageBinder) Resolve(l label.BuildLabel) (Resolver, string, error) {
	return p.ResolveF(l)
}

type PackageBind struct {
	Name   string
	Target func(Cfg) (Target, error)
}
type te struct {
	target Target
	error
}

func (p PackageBinder) Bind(name string, target func(Binder) (Target, error), transformation func(Cfg) (Cfg, error)) error {
	cache := map[[32]byte]Target{}
	p.Put(PackageBind{name, func(c Cfg) (Target, error) {
		if transformation != nil {
			var err error
			c, err = transformation(c)
			if err != nil {
				return Target{}, err
			}
		}
		h := sha256.New()
		err := c.Hash(h)
		if err != nil {
			return Target{}, err
		}
		s := [32]byte(h.Sum([]byte("pkg")))
		t, ok := cache[s]
		if ok {
			return t, nil
		}
		hb, _ := HashBinder{Inner: p, HName: name}.WithCfg(func(_ Cfg) (Cfg, error) {
			return c, nil
		})
		t, err = target(hb)
		if err != nil {
			return Target{}, err
		}
		cache[s] = t
		return t, nil
	}})
	return nil
}
func (p PackageBinder) WithCfg(transform func(Cfg) (Cfg, error)) (Binder, error) {
	t, err := transform(p.F)
	if err != nil {
		return nil, err
	}
	p.F = t
	return p, nil
}
func (p PackageBinder) Cfg() Cfg {
	return p.F
}
func (p PackageBinder) Name(name string) label.BuildLabel {
	l := p.Lab
	l.Name = name
	return l
}
func (p PackageBinder) Runner() sandbox.Runner {
	return p.R
}
func (p PackageBinder) Get(name string, c func(Cfg) (Cfg, error)) (Target, error) {
	var t Target
	s := ""
	for _, x := range strings.Split(name, "#") {
		if s != "" {
			s += "#"
		}
		s += x
		u := q.Get[PackageBind, te](p.Q, func(pb PackageBind) (te, bool) {
			if pb.Name == s {
				f, err := c(p.Cfg())
				if err != nil {
					return te{Target{}, err}, true
				}
				t, err := pb.Target(f)
				return te{t, err}, true
			}
			return te{}, false
		})
		if u.error != nil {
			return Target{}, u.error
		}
		t = u.target
	}
	return t, nil
}

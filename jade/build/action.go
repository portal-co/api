package build

import (
	"fmt"
	"path"

	"portal.pc/jade/erroror"
	"portal.pc/jade/file"
	"portal.pc/jade/sandbox"
	"portal.pc/jade/state"
	"golang.org/x/sync/errgroup"
)

type Result struct {
	Files map[string]file.File
}
type Action struct {
	bindBase func(Result) (Action, error)
	Deps     map[string]struct {
		Act Action
	}
	CmdAct *struct {
		Cmd   []string
		Ninja bool
		Name  string
		Outs  []string
	}
	Pure *erroror.ErrorOr[Result]
}

func (a Action) Bind(f func(Result) (Action, error)) Action {
	if a.bindBase == nil {
		a.bindBase = f
		return a
	}
	o := a.bindBase
	a.bindBase = func(r Result) (Action, error) {
		p, err := o(r)
		if err != nil {
			return Action{}, err
		}
		return p.Bind(f), nil
	}
	return a
}
func (a Action) buildBase(s state.State) (Result, error) {
	d := map[string]file.File{}
	var g errgroup.Group
	for k, v := range a.Deps {
		k := k
		v := v
		g.Go(func() error {
			w, err := v.Act.Build(s)
			if err != nil {
				return err
			}
			for k2, v2 := range w.Files {
				d[path.Join(k, k2)] = v2
			}
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return Result{}, err
	}
	if a.CmdAct != nil {
		f, err := sandbox.Run(s, d, a.CmdAct.Cmd, a.CmdAct.Ninja, a.CmdAct.Name, a.CmdAct.Outs)
		if err != nil {
			return Result{}, err
		}
		return Result{Files: f}, nil
	}
	if a.Pure != nil {
		return a.Pure.Run()
	}
	return Result{}, fmt.Errorf("invalid action")
}
func (a Action) Build(s state.State) (Result, error) {
	t, err := a.buildBase(s)
	if err != nil {
		return Result{}, err
	}
	if a.bindBase == nil {
		return t, nil
	}
	u, err := a.bindBase(t)
	if err != nil {
		return Result{}, err
	}
	return u.Build(s)
}

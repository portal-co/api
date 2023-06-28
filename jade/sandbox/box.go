package sandbox

import (
	"fmt"
	"os"
	"os/exec"

	"portal.pc/jade/file"
	"portal.pc/jade/ninja"
	"portal.pc/jade/state"
	"golang.org/x/sync/errgroup"
)

func Run(state state.State, inputs map[string]file.File, cmd []string, xninja bool, name string, outs []string) (map[string]file.File, error) {
	ix := map[string]string{}
	var g errgroup.Group
	for k, v := range inputs {
		k := k
		v := v
		g.Go(func() error {
			a, err := v.ToFs()
			if err != nil {
				return err
			}
			ix[k] = a
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return nil, err
	}
	t, err := os.MkdirTemp("/tmp", "prtl-sbox-*")
	if err != nil {
		return nil, err
	}
	// g = errgroup.Group{}
	for k, v := range ix {
		err := os.Symlink(v, t+"/"+k)
		if err != nil {
			return nil, err
		}
	}
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Dir = t
	err = state.Spawn(name, func() error {
		return c.Run()
	})
	if err != nil {
		return nil, err
	}
	if xninja {
		g = errgroup.Group{}
		for _, o := range outs {
			o := o
			g.Go(func() error {
				return ninja.DoNinja(state, t, o)
			})
		}
		err = g.Wait()
		if err != nil {
			return nil, err
		}
	}
	r, err := state.AddDir(t)
	if err != nil {
		return nil, err
	}
	o := map[string]file.File{}
	for _, p := range outs {
		o[p] = file.File{Ipfs: fmt.Sprintf("%s/%s", r, p)}
	}
	return o, nil
}

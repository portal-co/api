package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	hos "github.com/hack-pad/hackpadfs/os"
	icore "github.com/ipfs/boxo/coreiface"
	"go.uber.org/fx"

	"github.com/moby/sys/mountinfo"
	"github.com/portal-co/api/pkg/fxipfs"
	"github.com/portal-co/remount"
)

type State struct {
	remount.I
	Main string
}

var CmdState = fx.Provide(func(i icore.CoreAPI, m fxipfs.MountPath) Runner {
	return State{remount.I{i}, string(m)}
})

func (state State) Run(inputs map[string]string, cmd []string, outs []string) (string, error) {
	ix := inputs
	t, err := os.MkdirTemp("/tmp", "prtl-sbox-base-*")
	if err != nil {
		return "", err
	}
	// g = errgroup.Group{}
	l := []string{}
	for k, v := range ix {
		err := os.Symlink(path.Join(state.Main, v), path.Join(t, k))
		if err != nil {
			return "", err
		}
		l = append(l, path.Join(t, k))
	}
	u, err := os.MkdirTemp("/tmp", "prtl-sbox-8")
	if err != nil {
		return "", err
	}
	d := exec.Command("/usr/bin/env", "bindfs", "-f", "--resolve-symlinks", "-p", "a+Xx", t, u)
	d.Stderr = os.Stderr
	err = d.Start()
	if err != nil {
		err = fmt.Errorf("binding at %s at %w", u, err)
		return "", err
	}
	defer func() {
		err := d.Process.Signal(syscall.SIGINT)
		if err != nil {
			return
		}
		err = d.Wait()
		if err != nil {
			err = fmt.Errorf("binding at %s at %w", u, err)
			return
		}
		// d := exec.Command("fusermount", "-u", u)
		// d.Stderr = os.Stderr
		// if err != nil {
		// 	return
		// }
		// err = d.Run()
	}()
	b := false
	for !b {
		b, err = mountinfo.Mounted(u)
		if err != nil {
			return "", err
		}
	}
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = t
	err = c.Run()
	if err != nil {
		return "", err
	}
	return remount.Push(state.I, hos.NewFS(), t[1:])
}

type Runner interface {
	Run(inputs map[string]string, cmd []string, outs []string) (string, error)
}

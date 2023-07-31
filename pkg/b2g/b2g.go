package b2g

import (
	"fmt"
	"io"
	"sync"

	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"github.com/portal-co/api/pkg/gen"
	"github.com/portal-co/remount"
	"golang.org/x/sync/errgroup"
)

type Buck2 struct {
}

func (b Buck2) Gen() gen.Generator {
	return func(i remount.I, p string, cfg gen.GenCfg) (string, error) {
		m, err := mem.NewFS()
		if err != nil {
			return "", err
		}
		c, err := hackpadfs.Create(m, ".buckconfig")
		if err != nil {
			return "", err
		}
		var mtx sync.Mutex
		var g errgroup.Group
		for k, l := range cfg.Cells {
			k := k
			l := l
			g.Go(func() error {
				o, err := i.Open(l + "/.buckconfig")
				if err != nil {
					return err
				}
				defer o.Close()
				mtx.Lock()
				defer mtx.Unlock()
				_, err = fmt.Fprintf(remount.B{c}, "[repositories]\n%s = %s/%s\n", k, p, l)
				if err != nil {
					return err
				}
				_, err = io.Copy(remount.B{c}, o)
				return err
			})
		}
		err = g.Wait()
		if err != nil {
			return "", err
		}
		c.Close()
		return remount.Push(i, m, ".")
	}
}

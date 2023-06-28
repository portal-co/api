package cache

import (
	"io/fs"
	"os"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"portal.pc/redav"
	"github.com/tebeka/atexit"
	"golang.org/x/sync/errgroup"
)

var cache map[fs.FS]string

var global errgroup.Group

func init() {
	atexit.Register(func() {
		global.Wait()
	})
}

func GetFS(f fs.FS) (string, error) {
	s, ok := cache[f]
	if ok {
		return s, nil
	}
	g := redav.FuseFS(redav.Dir{f})
	t, err := os.MkdirTemp("/tmp", "fscache-*")
	if err != nil {
		return "", err
	}
	c, err := fuse.Mount(t)
	if err != nil {
		return "", err
	}
	global.Go(func() error {
		defer c.Close()
		return fusefs.Serve(c, g)
	})
	cache[f] = t
	return t, nil
}

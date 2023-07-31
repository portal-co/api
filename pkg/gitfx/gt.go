package gitfx

import (
	"fmt"
	"sync"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"go.uber.org/fx"
)

type Cloner func(p string, s storage.Storer) (*git.Repository, error)

func Basic(m transport.AuthMethod, prefix string) Cloner {
	cache := map[string]*git.Repository{}
	var mx sync.Mutex
	return func(p string, s storage.Storer) (*git.Repository, error) {
		mx.Lock()
		defer mx.Unlock()
		x, ok := cache[p]
		if ok {
			w, err := x.Worktree()
			if err != nil {
				return nil, err
			}
			err = w.Pull(&git.PullOptions{Auth: m})
			return x, err
		}
		x, err := git.Clone(s, memfs.New(), &git.CloneOptions{
			URL: fmt.Sprintf(prefix, p),
			// Auth: transport.,
			Auth: m,
		})
		if err == nil {
			cache[p] = x
		}
		return x, err
	}
}
func Gh(key string) Cloner {
	return Basic(&http.BasicAuth{
		Username: "_",
		Password: key,
	}, "https://github.com/%s")
}
func AsCloner(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"cloner"`),
	)
}

type Cloners struct {
	fx.In
	Vals []Cloner `group:"cloner"`
}

func (c Cloners) Clone(p string, s storage.Storer) (*git.Repository, error) {
	for _, x := range c.Vals {
		y, err := x(p, s)
		if err == nil {
			return y, nil
		}
	}
	return nil, fmt.Errorf("no repository found")
}

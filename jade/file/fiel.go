package file

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"plugin"

	"portal.pc/jade/cache"
	"github.com/hairyhenderson/go-fsimpl/gitfs"
)

type File struct {
	Ipfs      string
	GitObject *struct {
		Name   string
		Sub    string
		Commit string
		In     File
	}
	PluginFS *struct {
		In   File
		Name string
		Sub  string
		Arg  any
	}
}

func (f File) ToFs() (string, error) {
	if f.Ipfs != "" {
		return fmt.Sprintf("/ipfs/%s", f.Ipfs), nil
	}
	if f.GitObject != nil {
		g, err := f.GitObject.In.ToFs()
		if err != nil {
			return "", err
		}
		u, err := url.Parse(fmt.Sprintf("file://%s#%s", g, f.GitObject.Commit))
		if err != nil {
			return "", err
		}
		h, err := gitfs.New(u)
		if err != nil {
			return "", err
		}
		ff, err := cache.GetFS(h)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s/%s", ff, f.GitObject.Sub), nil
	}
	if f.PluginFS != nil {
		g, err := f.PluginFS.In.ToFs()
		if err != nil {
			return "", err
		}
		p, err := plugin.Open(fmt.Sprintf("%s/%s", g, f.PluginFS.Name))
		if err != nil {
			return "", err
		}
		q, err := p.Lookup("PluginEntry")
		if err != nil {
			return "", err
		}
		s, err := q.(func(any) (fs.FS, error))(f.PluginFS.Arg)
		if err != nil {
			return "", err
		}
		ff, err := cache.GetFS(s)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s/%s", ff, f.PluginFS.Sub), nil
	}
	return "", fmt.Errorf("invalid file")
}
func (f File) Sub(s string) File {
	if f.Ipfs != "" {
		f.Ipfs += "/" + s
	}
	if f.GitObject != nil {
		f.GitObject.Sub += "/" + s
	}
	if f.PluginFS != nil {
		f.PluginFS.Sub += "/" + s
	}
	return f
}
func (f File) Hash() ([]byte, error) {
	p, err := f.ToFs()
	if err != nil {
		return nil, err
	}
	o, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer o.Close()
	h := sha256.New()
	_, err = io.Copy(h, o)
	if err != nil {
		return nil, err
	}
	return h.Sum([]byte{}), nil
}

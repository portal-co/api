package build

import (
	"encoding/gob"
	"hash"

	"go.starlark.net/starlark"
)

type Cfg struct {
	*starlark.Dict
}

func (c Cfg) Hash(h hash.Hash) error {
	for _, k := range c.Keys() {
		s, _ := k.Hash()
		err := gob.NewEncoder(h).Encode(s)
		if err != nil {
			return err
		}
		x, _, _ := c.Get(k)
		err = gob.NewEncoder(h).Encode(x)
		if err != nil {
			return err
		}
	}
	return nil
}

package state

import (
	shell "github.com/ipfs/go-ipfs-api"
)

type Pool interface {
	Spawn(msg string, fn func() error) error
}
type State struct {
	Pool
	*shell.Shell
}

package gen

import "github.com/portal-co/remount"

type Generator func(i remount.I, p string, cfg GenCfg) (string, error)
type GenCfg struct {
	Cells map[string]string
}

package ninja

import (
	"fmt"
	"os/exec"
	"strings"

	"portal.pc/jade/state"
	"golang.org/x/sync/errgroup"
)

func Ninja(a []string, d string) string {
	n := exec.Command("ninja", a...)
	n.Dir = d
	x, _ := n.CombinedOutput()
	return string(x)
}
func Targets(d string) []string {
	return strings.Split(Ninja([]string{"-t", "targets", "depth", "1"}, d), "\n")
}
func Cmd(x, d string) string {
	s := strings.Split(Ninja([]string{"-t", "commands", x}, d), "\n")
	return s[len(s)-1]
}
func Inputs(x, d string) []string {
	return strings.Split(Ninja([]string{"-t", "inputs", x}, d), "\n")
}
func DoNinja(s state.State, d, x string) error {
	i := Inputs(x, d)
	var g errgroup.Group
	for _, j := range i {
		j := j
		g.Go(func() error {
			return DoNinja(s, d, j)
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}
	c := Cmd(x, d)
	m := exec.Command("/bin/sh", "-c", c)
	err = s.Spawn(fmt.Sprintf("Building %s", x), func() error {
		return m.Run()
	})
	if err != nil {
		return err
	}
	return nil
}

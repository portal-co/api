package build

func CCCompile(compiler Action, hdrs Action, hdrOuts []string, srcs Action, srcOuts []string, lang string) Action {
	return Action{
		Deps: map[string]struct{ Act Action }{
			"cc": struct{ Act Action }{Act: compiler},
			"pps": struct{ Act Action }{Act: Action{
				Deps: map[string]struct{ Act Action }{
					"cc":   struct{ Act Action }{Act: compiler},
					"srcs": struct{ Act Action }{Act: srcs},
					"hdrs": struct{ Act Action }{Act: hdrs},
				},
				CmdAct: &struct {
					Cmd   []string
					Ninja bool
					Name  string
					Outs  []string
				}{
					Cmd:  []string{"/bin/sh", "-c", "find ./srcs -exec './cc -x" + lang + " -I./hdrs -E {} -o ./out/{}'"},
					Outs: []string{"./out"},
				},
			}},
		},
		CmdAct: &struct {
			Cmd   []string
			Ninja bool
			Name  string
			Outs  []string
		}{
			Cmd:  []string{"/bin/sh", "-c", "./cc -x" + lang + " $(find ./pps) -Wl,-r -o out.o"},
			Outs: []string{"out.o"},
		},
	}
}

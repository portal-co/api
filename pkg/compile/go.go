package compile

import (
	"encoding/base64"
	"fmt"

	"github.com/portal-co/api/pkg/sandbox"
)

func GoCompile(st sandbox.Runner, compiler, src, pkg string, imap map[string]string) (string, error) {
	d := map[string]string{}
	im := ""
	d["go"] = compiler
	d["src"] = src
	for k, v := range imap {
		d["im/"+k] = v
		im += fmt.Sprintf("\npackagefile %s=im/%s", k, k)
	}
	s, err := st.Run(d, []string{"sh", "-c", fmt.Sprintf("echo %s | base64 -d > importcfg;./go/tool/compile -o obj.o -p %s -importcfg ./importcfg $(find ./src)", base64.StdEncoding.EncodeToString([]byte(im)), pkg)}, []string{"obj.a"})
	return s + "/obj.a", err
}

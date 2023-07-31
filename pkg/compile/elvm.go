package compile

import (
	"encoding/base64"
	"fmt"

	"github.com/portal-co/api/pkg/sandbox"
)

type ShimCtx struct {
	W2C2, ELVM, RISCW string
}

type ShimCCToolchain interface {
	CC(st sandbox.Runner, in string) (string, error)
}

func WasmToElvm(st sandbox.Runner, in string, ctx ShimCtx) (string, error) {
	w, err := st.Run(map[string]string{"target.wasm": in, "w2c2_bundle": ctx.W2C2}, []string{"./w2c2_bundle/w2c2", "target.wasm", "target.c"}, []string{"target.c"})
	if err != nil {
		return "", err
	}
	c, err := st.Run(map[string]string{"t": w, "w2c2_bundle": ctx.W2C2}, []string{"/bin/sh", "-c", fmt.Sprintf("cat ./w2c2_bundle/w2c2_base.h ./t/* > ./target.c;echo %s | base64 -d >> target;c", base64.StdEncoding.EncodeToString([]byte(`
	void boot__putchar(void *_,U32 c){
		putchar(c);
	}
	U32 boot__getchar(void *_){
		return getchar();
	}
	int main(){
		return target__main();
	}
	`)))}, []string{"target.c"})
	if err != nil {
		return "", err
	}
	return st.Run(map[string]string{"target.c": c + "/target.c", "elvm": ctx.ELVM}, []string{"elvm/8cc", "target.c", "-o", "target.eir"}, []string{"target.eir"})
}
func RiscW(st sandbox.Runner, cc ShimCCToolchain, in, wit string, ctx ShimCtx) (string, error) {
	x, err := st.Run(map[string]string{"in.rv.exe": in, "in.wit": wit, "w": ctx.RISCW}, []string{"w", "in.rv.exe", "in.wit", "-o", "out.cc"}, []string{"out.cc"})
	if err != nil {
		return "", err
	}
	return cc.CC(st, x+"/out.cc")
}

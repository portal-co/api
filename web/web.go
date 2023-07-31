package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	p2phttp "github.com/libp2p/go-libp2p-http"
	"github.com/portal-co/api/pkg/discfx"
	"github.com/portal-co/api/pkg/fxgh"
	"github.com/portal-co/api/pkg/fxipfs"
	"github.com/portal-co/api/pkg/httpfx"
	"github.com/portal-co/api/pkg/p2pfx"
	"github.com/portal-co/api/pkg/sandbox"

	icore "github.com/ipfs/boxo/coreiface"
	"go.uber.org/fx"
)

func main() {
	// c := fxipfs.CoreAPI
	// rp, err := fsrepo.Open("./portal")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ig := &node.BuildCfg{Online: true}
	// if os.Getenv("REMOTE_TARGET") != "" {
	// 	c = fxipfs.RemoteCoreAPI(os.Getenv("REMOTE_TARGET"))
	// }
	ghk := os.Getenv("GH_KEY")
	p := os.Getenv("PORT")
	if p == "" {
		p = "9091"
	}
	a := fx.New(p2pfx.Host, httpfx.HTTPServeOpt(":"+p), httpfx.HTTPServeOptHost(p2phttp.DefaultP2PProtocol), httpfx.HTTP, fxipfs.Node, fxipfs.CoreAPI, fxipfs.Cfg, fxipfs.Mount, fxgh.Events(ghk), discfx.Client(os.Getenv("DISCORD_TOKEN")), sandbox.CmdState, fx.Invoke(func(n icore.CoreAPI, m fxipfs.MountPath, s *discordgo.Session, r sandbox.Runner) {
		a, err := r.Run(map[string]string{}, []string{"/usr/bin/env", "touch", "a"}, []string{"a"})
		if err != nil {
			fmt.Print(err)
		}
		fmt.Print(a)
		b, err := r.Run(map[string]string{"a": a}, []string{"/usr/bin/env", "find", "."}, []string{"a"})
		if err != nil {
			fmt.Print(err)
		}
		fmt.Print(b)
	}))
	a.Run()
}

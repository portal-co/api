package fxgh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
	"github.com/portal-co/api/pkg/httpfx"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
)

type GhOption func(*githubevents.EventHandler) error

func Events(key string) fx.Option {
	return fx.Options(fx.Provide(fx.Annotate(func(x []GhOption) (*githubevents.EventHandler, error) {
		h := githubevents.New(key)
		for _, y := range x {
			err := y(h)
			if err != nil {
				return nil, err
			}
		}
		return h, nil
	}, fx.ParamTags(`group:"github_hooks"`))), fx.Provide(httpfx.AsRoute(func(h *githubevents.EventHandler) httpfx.Route {
		return httpfx.RouteFunc("/github", func(w http.ResponseWriter, r *http.Request) {
			err := h.HandleEventRequest(r)
			if err != nil {
				fmt.Fprintln(w, err)
			}
		})
	})), fx.Provide(func() *github.Client {
		return github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: key})))
	}))
}
func AsHook(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"github_hook"`),
	)
}

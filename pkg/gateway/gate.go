package gateway

import (
	"net/http"

	icore "github.com/ipfs/boxo/coreiface"
	"github.com/portal-co/api/pkg/httpfx"
	"github.com/portal-co/remount"
)

func NewGateway(x icore.CoreAPI) httpfx.Route {
	return httpfx.Route{Pattern: "/ipfs/", Handler: http.FileServer(http.FS(remount.I{x}))}
}

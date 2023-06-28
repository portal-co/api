package hr2

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Routes func(string) (chi.Router, bool)

func (hr Routes) Match(rctx *chi.Context, method, path string) bool {
	return true
}

func (hr Routes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := requestHost(r)
	if router, ok := hr(strings.ToLower(host)); ok {
		router.ServeHTTP(w, r)
		return
	}
	if router, ok := hr(strings.ToLower(getWildcardHost(host))); ok {
		router.ServeHTTP(w, r)
		return
	}
	if router, ok := hr("*"); ok {
		router.ServeHTTP(w, r)
		return
	}
	if router, ok := hr(""); ok {
		router.ServeHTTP(w, r)
		return
	}
	http.Error(w, http.StatusText(404), 404)
}

func (hr Routes) Routes() []chi.Route {
	x, _ := hr("")
	return x.Routes()
}

func (hr Routes) Middlewares() chi.Middlewares {
	return chi.Middlewares{}
}

func requestHost(r *http.Request) (host string) {
	// not standard, but most popular
	host = r.Header.Get("X-Forwarded-Host")
	if host != "" {
		return
	}

	// RFC 7239
	host = r.Header.Get("Forwarded")
	_, _, host = parseForwarded(host)
	if host != "" {
		return
	}

	// if all else fails fall back to request host
	host = r.Host
	return
}

func parseForwarded(forwarded string) (addr, proto, host string) {
	if forwarded == "" {
		return
	}
	for _, forwardedPair := range strings.Split(forwarded, ";") {
		if tv := strings.SplitN(forwardedPair, "=", 2); len(tv) == 2 {
			token, value := tv[0], tv[1]
			token = strings.TrimSpace(token)
			value = strings.TrimSpace(strings.Trim(value, `"`))
			switch strings.ToLower(token) {
			case "for":
				addr = value
			case "proto":
				proto = value
			case "host":
				host = value
			}

		}
	}
	return
}

func getWildcardHost(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) > 1 {
		wildcard := append([]string{"*"}, parts[1:]...)
		return strings.Join(wildcard, ".")
	}
	return strings.Join(parts, ".")
}

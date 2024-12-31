package caddy_analytics

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func parseCaddyfileHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	a := defaultConfig()
	h.Dispenser.Next()

	if !h.Dispenser.Args(&a.Provider) {
		return nil, h.ArgErr()
	}

	for nesting := h.Dispenser.Nesting(); h.Dispenser.NextBlock(nesting); {
		switch h.Dispenser.Val() {
		case "host":
			if !h.Dispenser.NextArg() {
				return nil, h.ArgErr()
			}
			a.Server = h.Dispenser.Val()
		case "admin_token":
			if !h.Dispenser.NextArg() {
				return nil, h.ArgErr()
			}
			a.AdminToken = h.Dispenser.Val()
		default:
			return nil, h.Errf("unrecognized subdirective: %s", h.Dispenser.Val())
		}
	}

	return a, nil
}

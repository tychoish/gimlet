//go:build go1.11
// +build go1.11

package gimlet

import (
	"net/http/httputil"

	"github.com/tychoish/grip"
	"github.com/tychoish/grip/message"
	"github.com/tychoish/grip/send"
)

// Proxy adds a simple reverse proxy handler to the specified route,
// based on the options described in the ProxyOption structure.
// In most cases you'll want to specify a route matching pattern
// that captures all routes that begin with a specific prefix.
func (r *APIRoute) Proxy(opts ProxyOptions) *APIRoute {
	if err := opts.Validate(); err != nil {
		grip.Alert(message.WrapError(err, message.Fields{
			"message":          "invalid proxy options",
			"route":            r.route,
			"version":          r.version,
			"existing_handler": r.handler != nil,
		}))
		return r
	}

	r.handler = (&httputil.ReverseProxy{
		Transport:    opts.Transport,
		ErrorLog:     send.MakeStandard(grip.Sender()),
		Director:     opts.director,
		ErrorHandler: opts.ErrorHandler,
	}).ServeHTTP

	return r
}

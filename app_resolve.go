package gimlet

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tychoish/grip"
	"github.com/tychoish/grip/recovery"
	"github.com/urfave/negroni"
)

// Handler returns a handler interface for integration with other
// server frameworks.
func (a *APIApp) Handler() (http.Handler, error) {
	switch a.routerImpl {
	case RouterImplGorilla:
		return a.getNegroni()
	case RouterImplChi:
		return a.getChi()
	default:
		return nil, errors.Errorf("invalid router specified")
	}
}

// Resolve processes the data in an application instance, including
// all routes and creats a mux.Router object for the application
// instance.
func (a *APIApp) Resolve() error {
	if a.isResolved {
		return nil
	}

	if err := a.routerImpl.Validate(); err != nil {
		return errors.Wrap(err, "improperly specified router implementation")
	}

	catcher := grip.NewBasicCatcher()
	for m := range iterMerge(a.middleware, a.wrappers) {
		switch mw := m.(type) {
		case HandlerFuncWrapper:
			continue
		case HandlerWrapper:
			continue
		case Middleware:
			continue
		default:
			catcher.Errorf("middleware of type %T is not supported as middleware", mw)
		}
	}

	if catcher.HasErrors() {
		return catcher.Resolve()
	}

	switch a.routerImpl {
	case RouterImplGorilla:
		if a.router == nil {
			a.router = mux.NewRouter().StrictSlash(a.StrictSlash)
		}

		if err := a.attachRoutes(a.router, true); err != nil {
			return errors.WithStack(err)
		}

		a.isResolved = true
	case RouterImplChi:
		if a.mux == nil {
			a.mux = chi.NewMux()
			if !a.StrictSlash {
				a.mux.Use(middleware.StripSlashes)
			}
		}

		if err := a.attachRoutes(a.mux, true); err != nil {
			return errors.WithStack(err)
		}
		a.isResolved = true
	}

	return nil
}

func iterMerge(sl ...[]interface{}) <-chan interface{} {
	out := make(chan interface{}, len(sl))
	go func() {
		defer recovery.LogStackTraceAndContinue("problem merging iterators")
		defer close(out)
		for slidx := range sl {
			for idx := range sl[slidx] {
				out <- sl[slidx][idx]
			}
		}
	}()
	return out
}

// getHander internal helper resolves the negorni middleware for the
// application and returns it in the form of a http.Handler for use in
// stitching together applications.
func (a *APIApp) getNegroni() (*negroni.Negroni, error) {
	if err := a.Resolve(); err != nil {
		return nil, err
	}

	n := negroni.New()
	for _, m := range a.middleware {
		switch mw := m.(type) {
		case Middleware:
			n.Use(mw)
		case HandlerFuncWrapper:
			n.Use(WrapperMiddleware(mw))
		case HandlerWrapper:
			n.Use(WrapperHandlerMiddleware(mw))
		}
	}

	n.UseHandler(a.router)

	return n, nil
}

func (a *APIApp) getChi() (*chi.Mux, error) {
	a.mux = chi.NewMux()
	if !a.StrictSlash {
		a.mux.Use(middleware.StripSlashes)
	}

	for _, m := range a.middleware {
		switch mw := m.(type) {
		case Middleware:
			a.mux.Use(MiddlewareFunc(mw))
		case HandlerFuncWrapper:
			a.mux.Use(MiddlewareFunc(mw))
		case HandlerWrapper:
			a.mux.Use(mw)
		}
	}

	if err := a.Resolve(); err != nil {
		return nil, err
	}

	return a.mux, nil
}

func (a *APIApp) attachRoutes(muxer interface{}, addAppPrefix bool) error {
	if router, ok := muxer.(*mux.Router); ok {
		router.StrictSlash(a.StrictSlash)
	}

	catcher := grip.NewCatcher()
	for _, route := range a.routes {
		if !route.IsValid() {
			catcher.Errorf("%s is not a valid route, skipping", route)
			continue
		}

		var methods []string
		for _, m := range route.methods {
			methods = append(methods, strings.ToLower(m.String()))
		}

		routeString := ""
		if route.version >= 0 {
			routeString = route.resolveVersionedRoute(a, addAppPrefix)
		} else if a.NoVersions {
			routeString = route.resolveLegacyRoute(a, addAppPrefix)
		} else {
			catcher.Errorf("skipping '%s', because of versioning error", route)
			continue
		}

		var invalidMiddleware bool
		for idx := range route.wrappers {
			switch mw := route.wrappers[idx].(type) {
			case HandlerWrapper:
				continue
			case HandlerFuncWrapper:
				continue
			case Middleware:
				continue
			default:
				invalidMiddleware = true
				catcher.Errorf("mw#%d for %s is %T which is not supported", idx, route.String(), mw)
			}
		}

		if invalidMiddleware {
			continue
		}

		switch router := muxer.(type) {
		case *mux.Router:
			handler := route.getHandlerWithMiddlware(a.wrappers)
			if route.isPrefix {
				router.PathPrefix(routeString).Handler(handler).Methods(methods...)
			} else {
				router.Handle(routeString, handler).Methods(methods...)
			}
		case *chi.Mux:
			mws := route.getMiddlewareSlice(a.wrappers)

			if len(mws) == 0 {
				for idx := range methods {
					router.Method(methods[idx], routeString, route.handler)
				}
			} else {
				for idx := range methods {
					router.With(mws...).Method(methods[idx], routeString, route.handler)
				}
			}

		default:
			return errors.Errorf("%T is not a valid mux type", muxer)
		}
	}

	return catcher.Resolve()
}

func (r *APIRoute) getRoutePrefix(app *APIApp, addAppPrefix bool) string {
	if !addAppPrefix {
		return ""
	}

	if r.overrideAppPrefix && r.prefix != "" {
		return r.prefix
	}

	return app.prefix
}

func (r *APIRoute) resolveLegacyRoute(app *APIApp, addAppPrefix bool) string {
	var output string

	prefix := r.getRoutePrefix(app, addAppPrefix)

	if prefix != "" {
		output += prefix
	}

	if r.prefix != prefix && r.prefix != "" {
		output += r.prefix
	}

	output += r.route

	return output
}

func (r *APIRoute) getVersionPart(app *APIApp) string {
	var versionPrefix string

	if !app.SimpleVersions {
		versionPrefix = "v"
	}

	return fmt.Sprintf("/%s%d", versionPrefix, r.version)
}

func (r *APIRoute) resolveVersionedRoute(app *APIApp, addAppPrefix bool) string {
	var (
		output string
		route  string
	)

	route = r.route
	firstPrefix := r.getRoutePrefix(app, addAppPrefix)

	if firstPrefix != "" {
		output += firstPrefix
	}

	output += r.getVersionPart(app)

	if r.prefix != firstPrefix && r.prefix != "" {
		output += r.prefix
	}

	output += route

	return output
}

func (r *APIRoute) getHandlerWithMiddlware(mws []interface{}) http.Handler {
	if len(mws) == 0 && len(r.wrappers) == 0 {
		return r.handler
	}

	n := negroni.New()

	for m := range iterMerge(mws, r.wrappers) {
		switch mw := m.(type) {
		case HandlerFuncWrapper:
			n.Use(WrapperMiddleware(mw))
		case HandlerWrapper:
			n.Use(mw)
		case Middleware:
			n.Use(mw)
		}
	}

	n.UseHandler(r.handler)
	return n
}

func (r *APIRoute) getMiddlewareSlice(mws []interface{}) chi.Middlewares {
	out := make(chi.Middlewares, 0, len(mws)+len(r.wrappers))

	for m := range iterMerge(mws, r.wrappers) {
		switch mw := m.(type) {
		case HandlerWrapper:
			out = append(out, mw)
		case HandlerFuncWrapper:
			out = append(out, MiddlewareFunc(mw))
		case Middleware:
			out = append(out, MiddlewareFunc(mw))
		}
	}

	return out
}

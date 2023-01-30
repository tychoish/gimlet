package gimlet

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tychoish/fun/erc"
	"github.com/urfave/negroni"
)

// AssembleHandler takes a router and one or more applications and
// returns an application.
//
// Eventually the router will become an implementation detail of
// this/related functions.
func AssembleHandlerGorilla(router *mux.Router, apps ...*APIApp) (http.Handler, error) {
	catcher := &erc.Collector{}
	mws := []interface{}{}

	seenPrefixes := make(map[string]struct{})

	for _, app := range apps {
		if app.prefix != "" {
			if _, ok := seenPrefixes[app.prefix]; ok {
				catcher.Add(errors.Errorf("route prefix '%s' defined more than once", app.prefix))
			}
			seenPrefixes[app.prefix] = struct{}{}

			r := router.PathPrefix(app.prefix).Subrouter()
			catcher.Add(app.attachRoutes(r, false)) // this adds wrapper middlware
			router.PathPrefix(app.prefix).Handler(buildNegroni(r, app.middleware...))
		} else {
			mws = append(mws, app.middleware...)
			catcher.Add(app.attachRoutes(router, true))
		}
	}

	if catcher.HasErrors() {
		return nil, catcher.Resolve()
	}

	return buildNegroni(router, mws), nil
}

func AssembleHandlerChi(router *chi.Mux, apps ...*APIApp) (out http.Handler, err error) {
	out = router
	catcher := &erc.Collector{}
	mws := []interface{}{}

	seenPrefixes := make(map[string]struct{})

	defer func() {
		if p := recover(); p != nil {
			catcher.Add(fmt.Errorf("chi.Mux encountered error: %+v", p))
		}
		err = catcher.Resolve()
		if catcher.HasErrors() {
			out = nil
		}
	}()

	router.Use(convertMidlewares(mws...)...)

	for _, app := range apps {
		if app.prefix != "" {
			if _, ok := seenPrefixes[app.prefix]; ok {
				catcher.Add(errors.Errorf("route prefix '%s' defined more than once", app.prefix))
			}
			seenPrefixes[app.prefix] = struct{}{}

			router.With(convertMidlewares(app.middleware...)...).Route(app.prefix, func(r chi.Router) {
				catcher.Add(app.attachRoutes(r, false)) // this adds wrapper middlware
			})
		} else {
			mws = append(mws, app.middleware...)
			catcher.Add(app.attachRoutes(router, true))
		}
	}

	return
}

func convertMidlewares(mws ...interface{}) []func(http.Handler) http.Handler {
	out := make([]func(http.Handler) http.Handler, 0, len(mws))

	for _, mw := range mws {
		switch m := mw.(type) {
		case HandlerWrapper:
			out = append(out, m)
		case HandlerFuncWrapper:
			out = append(out, MiddlewareFunc(m))
		case Middleware:
			out = append(out, MiddlewareFunc((m)))
		}
	}

	return out
}

func buildNegroni(router http.Handler, mws ...interface{}) *negroni.Negroni {
	n := negroni.New()

	for _, m := range mws {
		switch mw := m.(type) {
		case HandlerWrapper:
			n.Use(mw)
		case HandlerFuncWrapper:
			n.Use(WrapperMiddleware(mw))
		case Middleware:
			n.Use(mw)
		}
	}

	n.UseHandler(router)
	return n
}

// MergeApplications takes a number of gimlet applications and
// resolves them, returning an http.Handler.
func MergeApplications(apps ...*APIApp) (http.Handler, error) {
	if len(apps) == 0 {
		return nil, errors.New("must specify at least one application")
	}

	var impl RouterImplementation
	for idx, app := range apps {
		if idx == 0 || impl == RouterImplUndefined {
			impl = app.routerImpl
			continue
		}
		if impl != app.routerImpl {
			return nil, errors.Errorf("cannot merge applications: app #%d uses %s, and all apps must use %s",
				idx, app.routerImpl, impl)
		}
	}

	switch impl {
	case RouterImplChi:
		return AssembleHandlerChi(chi.NewMux(), apps...)
	case RouterImplGorilla:
		return AssembleHandlerGorilla(mux.NewRouter(), apps...)
	default:
		return nil, errors.New("undefined router implementation")
	}
}

// Merge takes multiple application instances and merges all of their
// routes into a single application.
//
// You must only call Merge once per base application, and you must
// pass more than one or more application to merge. Additionally, it
// is not possible to merge applications into a base application that
// has a prefix specified.
//
// When the merging application does not have a prefix, the merge
// operation will error if you attempt to merge applications that have
// duplicate cases. Similarly, you cannot merge multiple applications
// that have the same prefix: you should treat these errors as fatal.
func (a *APIApp) Merge(apps ...*APIApp) error {
	if a.prefix != "" {
		return errors.New("cannot merge applications into an application with a prefix")
	}

	if apps == nil {
		return errors.New("must specify apps to merge")
	}

	if a.hasMerged {
		return errors.New("can only call merge once per root application")
	}

	catcher := &erc.Collector{}
	seenPrefixes := make(map[string]struct{})

	for _, app := range apps {
		if app.prefix != "" {
			if _, ok := seenPrefixes[app.prefix]; ok {
				catcher.Add(fmt.Errorf("route prefix '%s' defined more than once", app.prefix))
			}
			seenPrefixes[app.prefix] = struct{}{}

			for _, route := range app.routes {
				r := a.PrefixRoute(app.prefix).Route(route.route).Version(route.version).Handler(route.handler)
				for _, m := range route.methods {
					r = r.Method(m.String())
				}

				r.overrideAppPrefix = route.overrideAppPrefix
				r.wrappers = append(app.middleware, route.wrappers...)
			}
		} else if app.middleware == nil {
			for _, r := range app.routes {
				if a.containsRoute(r.route, r.version, r.methods) {
					catcher.Add(fmt.Errorf("cannot merge route '%s' with existing application that already has this route defined", r.route))
				}
			}

			a.routes = append(a.routes, app.routes...)
		} else {
			for _, route := range app.routes {
				if a.containsRoute(route.route, route.version, route.methods) {
					catcher.Add(fmt.Errorf("cannot merge route '%s' with existing application that already has this route defined", route.route))
				}

				r := a.Route().Route(route.route).Version(route.version)
				for _, m := range route.methods {
					r = r.Method(m.String())
				}
				r.overrideAppPrefix = route.overrideAppPrefix
				r.wrappers = append(app.middleware, route.wrappers...)
			}
		}
	}

	a.hasMerged = true

	return catcher.Resolve()
}

func (a *APIApp) containsRoute(path string, version int, methods []httpMethod) bool {
	for _, r := range a.routes {
		if r.route == path && r.version == version {
			for _, m := range r.methods {
				for _, rm := range methods {
					if m == rm {
						return true
					}
				}
			}
		}
	}

	return false
}

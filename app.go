// Package gimlet is a toolkit for building JSON/HTTP interfaces (e.g. REST).
//
// Gimlet builds on standard library and common tools for building web
// applciations (e.g. Negroni and gorilla,) and is only concerned with
// JSON/HTTP interfaces, and omits support for aspects of HTTP
// applications outside of the scope of JSON APIs (e.g. templating,
// sessions.) Gimilet attempts to provide minimal convinences on top
// of great infrastucture so that your application can omit
// boilerplate and you don't have to build potentially redundant
// infrastructure.
package gimlet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cdr/grip"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/cors"
)

// WaitFunc is a function type returned by some functions that allows
// callers to wait on background processes started by the returning
// function.
//
// If the context passed to a wait function is canceled then the wait
// function should return immediately. You may wish to pass contexts
// with a different timeout to the wait function from the one you
// passed to the outer function to ensure correct waiting semantics.
type WaitFunc func(context.Context)

// APIApp is a structure representing a single API service.
type APIApp struct {
	StrictSlash    bool
	SimpleVersions bool
	NoVersions     bool
	isResolved     bool
	hasMerged      bool
	prefix         string
	port           int
	address        string
	routes         []*APIRoute

	routerImpl RouterImplementation
	router     *mux.Router
	mux        *chi.Mux
	middleware []interface{}
	wrappers   []interface{}
}

// RouterImplementation describes the http Routing infrastructure the
// Application will use to configure routes.
type RouterImplementation int

const (
	// RouterImplUndefined routers are not configured and it is an
	// error
	//
	// RouterImplGorilla uses Negroni and gorillia/mux for all
	// routing functions.
	//
	// RouterImplChi uses the go-chi/chi routing infrastructure.
	RouterImplUndefined = iota
	RouterImplGorilla
	RouterImplChi
)

// String implements fmt.Stringer for the RouterImplementation.
func (r RouterImplementation) String() string {
	switch r {
	case RouterImplGorilla:
		return "gorilla"
	case RouterImplChi:
		return "chi"
	case RouterImplUndefined:
		return "<undefined>"
	default:
		return "<invalid-router>"
	}

}

// Validate returns an error if the router implementation is invalid
// or not specified.
func (r RouterImplementation) Validate() error {
	switch r {
	case RouterImplChi, RouterImplGorilla:
		return nil
	case RouterImplUndefined:
		return errors.New("unspecified router implementation")
	default:
		return errors.Errorf("%d is not a valid router [%s]", r, r.String())
	}
}

// NewApp returns a pointer to an application instance. These
// instances have reasonable defaults and include middleware to:
// recover from panics in handlers, log information about the request,
// and gzip compress all data. Users must specify a default version
// for new methods.
func NewApp() *APIApp {
	a := &APIApp{
		port:        3000,
		StrictSlash: true,
		routerImpl:  RouterImplGorilla,
	}

	return a
}

// Router is the getter for an APIApp's router object. If the
// application isn't resolved or the app does not use the
// negroni/Gorilla mux stack, then the error return value is non-nil.
func (a *APIApp) Router() (*mux.Router, error) {
	if a.routerImpl != RouterImplGorilla {
		return nil, errors.Errorf("gorilla router is not configured [%s]", a.routerImpl)
	}

	if a.isResolved {
		return a.router, nil
	}
	return nil, errors.New("application is not resolved")
}

// Mux is a getter for an APIApp's chi Muxer object. If the
// application isn't resovled or uses a different routing stack, then
// this is an error.
func (a *APIApp) Mux() (*chi.Mux, error) {
	if a.routerImpl != RouterImplChi {
		return nil, errors.New("chi router is not configured")
	}

	if a.isResolved {
		return a.mux, nil
	}
	return nil, errors.New("application is not resolved")
}

// AddMiddleware adds a negroni handler as middleware to the end of
// the current list of middleware handlers.
//
// All Middleware is added before the router. If your middleware
// depends on executing within the context of the router/muxer, add it
// as a wrapper.
func (a *APIApp) AddMiddleware(m Middleware) *APIApp {
	a.middleware = append(a.middleware, m)
	return a
}

// AddMiddlewareFunc adds middleware in the form of a http.HandlerFunc
// wrapper.
//
// All Middleware is added before the router. If your middleware
// depends on executing within the context of the router/muxer, add it
// as a wrapper.
func (a *APIApp) AddMiddlewareFunc(m HandlerFuncWrapper) *APIApp {
	a.middleware = append(a.middleware, m)
	return a
}

// AddMiddlewareFunc adds middleware in the form of a http.HandlerFunc
// wrapper.
//
// All Middleware is added before the router. If your middleware
// depends on executing within the context of the router/muxer, add it
// as a wrapper.
func (a *APIApp) AddMiddlewareHandler(m HandlerWrapper) *APIApp {
	a.middleware = append(a.middleware, m)
	return a
}

// AddWrapper adds a negroni handler as a wrapper for a specific route.
//
// These wrappers execute in the context of the router/muxer. If your
// middleware does not need access to the muxer's state, add it as a
// middleware.
func (a *APIApp) AddWrapper(m Middleware) *APIApp {
	a.wrappers = append(a.wrappers, m)
	return a
}

// AddWrapperFunc adds middleware, defined as a HandlerFuncWrapper to
// routes.
//
// These wrappers execute in the context of the router/muxer. If your
// middleware does not need access to the muxer's state, add it as a
// middleware.
func (a *APIApp) AddWrapperFunc(m HandlerFuncWrapper) *APIApp {
	a.wrappers = append(a.wrappers, m)
	return a
}

// AddWrapperFunc adds middleware, defined as a HandlerWrapper to
// routes.
//
// These wrappers execute in the context of the router/muxer. If your
// middleware does not need access to the muxer's state, add it as a
// middleware.
func (a *APIApp) AddWrapperHandler(m HandlerWrapper) *APIApp {
	a.wrappers = append(a.wrappers, m)
	return a
}

// ResetMiddleware removes *all* middleware handlers from the current
// application.
func (a *APIApp) ResetMiddleware() {
	a.middleware = []interface{}{}
}

// SetRouter allows you to configure which underlying router
// infrastructure the Application will use. It is an error to merge
// two applications with different routing implementations.
func (a *APIApp) SetRouter(r RouterImplementation) *APIApp {
	a.routerImpl = r
	return a
}

// RestWrappers removes all route-specific middleware from the
// current application.
func (a *APIApp) RestWrappers() {
	a.wrappers = []interface{}{}
}

func (a *APIApp) AddCORS(opts cors.Options) *APIApp {
	c := cors.New(opts)
	a.AddMiddleware(c)
	return a
}

// Run configured API service on the configured port. Before running
// the application, Run also resolves any sub-apps, and adds all
// routes.
//
// If you cancel the context that you pass to run, the application
// will gracefully shutdown, and wait indefinitely until the
// application has returned. To get different waiting behavior use
// BackgroundRun.
func (a *APIApp) Run(ctx context.Context) error {
	wait, err := a.BackgroundRun(ctx)

	if err != nil {
		return errors.WithStack(err)
	}

	wait(context.Background())

	return nil
}

// BackgroundRun is a non-blocking form of Run that allows you to
// manage a service running in the background.
func (a *APIApp) BackgroundRun(ctx context.Context) (WaitFunc, error) {
	n, err := a.Handler()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	conf := ServerConfig{
		Handler: n,
		Address: fmt.Sprintf("%s:%d", a.address, a.port),
		Timeout: time.Minute,
		Info:    fmt.Sprintf("app with '%s' prefix", a.prefix),
	}

	srv, err := conf.Resolve()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	grip.Noticef("starting %s on: %s:%d", a.prefix, a.address, a.port)

	return srv.Run(ctx)
}

// SetPort allows users to configure a default port for the API
// service. Defaults to 3000, and return errors will refuse to set the
// port to something unreasonable.
func (a *APIApp) SetPort(port int) error {
	defaultPort := 3000

	if port == a.port {
		grip.Warningf("port is already set to %d", a.port)
	} else if port <= 0 {
		a.port = defaultPort
		return fmt.Errorf("%d is not a valid port numbaer, using %d", port, defaultPort)
	} else if port > 65535 {
		a.port = defaultPort
		return fmt.Errorf("port %d is too large, using default port (%d)", port, defaultPort)
	} else if port < 1024 {
		a.port = defaultPort
		return fmt.Errorf("port %d is too small, using default port (%d)", port, defaultPort)
	} else {
		a.port = port
	}

	return nil
}

// SetHost sets the hostname or address for the application to listen
// on. Errors after resolving the application. You do not need to set
// this, and if unset the application will listen on the specified
// port on all interfaces.
func (a *APIApp) SetHost(name string) error {
	if a.isResolved {
		return fmt.Errorf("cannot set host to '%s', after resolving. Host is still '%s'",
			name, a.address)
	}

	a.address = name

	return nil
}

// SetPrefix sets the route prefix, adding a leading slash, "/", if
// necessary.
func (a *APIApp) SetPrefix(p string) {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	a.prefix = p
}

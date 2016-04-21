// Gimlet is a toolkit for building JSON/HTTP interfaces (e.g. REST).
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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/tychoish/grip"
	"github.com/tylerb/graceful"
)

// A structure representing a single API service
type ApiApp struct {
	routes         []*ApiRoute
	defaultVersion int
	isResolved     bool
	router         *mux.Router
	address        string
	port           int
	strictSlash    bool
	middleware     []negroni.Handler
}

// Returns a pointer to an application instance. These instances have
// reasonable defaults and include middleware to: recover from panics
// in handlers, log information about the request, and gzip compress
// all data. Users must specify a default version for new methods.
func NewApp() *ApiApp {
	a := &ApiApp{
		defaultVersion: -1, // this is the same as having no version prepended to the path.
		port:           3000,
		strictSlash:    true,
	}

	a.AddMiddleware(negroni.NewRecovery())
	a.AddMiddleware(NewAppLogger())
	a.AddMiddleware(gzip.Gzip(gzip.DefaultCompression))

	return a
}

// Specifies a default version for the application. Default versions
// must be 0 (no version,) or larger.
func (self *ApiApp) SetDefaultVersion(version int) {
	if version < 0 {
		grip.Warningf("%d is not a valid version", version)
	} else {
		self.defaultVersion = version
		grip.Noticef("Set default api version to /v%d/", version)
	}
}

func (self *ApiApp) Router() (*mux.Router, error) {
	if self.isResolved {
		return self.router, nil
	} else {
		return self.router, errors.New("application is not resolved")
	}
}

// Take one app and add its routes to the current app. Errors if the
// current app is resolved. If the apps have different default
// versions set, the versions on the second app are explicitly set.
func (self *ApiApp) AddApp(app *ApiApp) error {
	// if we've already resolved then it has to be an error
	if self.isResolved {
		return errors.New("cannot merge an app into a resolved app.")
	}

	// this is a weird case, so worth a warning, but not worth exiting
	if app.isResolved {
		grip.Warningln("merging a resolved app into an unresolved app may be an error.",
			"Continuing cautiously.")
	}
	// this is incredibly straightforward, just add the added routes to our routes list.
	if app.defaultVersion == self.defaultVersion {
		self.routes = append(self.routes, app.routes...)
		return nil
	}

	// This makes sure that instance default versions are
	// respected in routes when merging instances. This covers the
	// case where you assemble v1 and v2 of an api in different
	// places in your code and want to merge them in later.
	for _, route := range app.routes {
		if route.version == 0 {
			route.Version(app.defaultVersion)
		}
		self.routes = append(self.routes, route)
	}

	return nil
}

// Sets the trailing slash behavior to pass to the `mux` layer. When
// `true`, routes with and without trailing slashes resolve to the
// same target. When `false`, the trailing slash is meaningful. The
// default value for Gimlet apps is `true`, and this method should be
// replaced with something more reasonable in the future.
func (self *ApiApp) SetStrictSlash(v bool) {
	self.strictSlash = v
}

// Adds a negroni handler as middleware to the end of the current list
// of middleware handlers.
func (self *ApiApp) AddMiddleware(m negroni.Handler) {
	self.middleware = append(self.middleware, m)
}

// Removes *all* middleware handlers from the current application.
func (self *ApiApp) ResetMiddleware() {
	self.middleware = []negroni.Handler{}
}

// Run configured API service on the configured port. If you Registers
// middlewear for gziped responses and graceful shutdown with a 10
// second timeout.
func (self *ApiApp) Run() error {
	var err error
	if self.isResolved == false {
		err = self.Resolve()
	}

	n := negroni.New()
	for _, m := range self.middleware {
		n.Use(m)
	}

	n.UseHandler(self.router)

	listenOn := strings.Join([]string{self.address, strconv.Itoa(self.port)}, ":")
	grip.Noticeln("starting app on:", listenOn)

	graceful.Run(listenOn, 10*time.Second, n)
	return err
}

// Allows user to configure a default port for the API
// service. Defaults to 3000, and return errors will refuse to set the
// port to something unreasonable.
func (self *ApiApp) SetPort(port int) (err error) {
	defaultPort := 3000

	if port == self.port {
		grip.Warningf("port is already set to %d", self.port)
		return
	} else if port <= 0 && self.port != defaultPort {
		err = fmt.Errorf("%d is not a valid port numbaer, using %d", port, defaultPort)
		self.port = defaultPort
		return
	} else if port > 65535 {
		err = fmt.Errorf("port %d is too large, using default port (%d)", port, defaultPort)
		self.port = defaultPort
		return
	} else if port < 1024 {
		err = fmt.Errorf("port %d is too small, using default port (%d)", port, defaultPort)
		self.port = defaultPort
		return
	} else {
		self.port = port
		return
	}
}

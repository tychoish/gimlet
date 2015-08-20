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
	"fmt"
	"strconv"
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
	port           int
	strictSlash    bool
}

// Returns a pointer to an application instance. These instances have
// reasonable defaults. Users must specify a default version for new
// methods.
func NewApp() *ApiApp {
	return &ApiApp{
		defaultVersion: -1, // this is the same as having no version prepended to the path.
		port:           3000,
		strictSlash:    true,
	}
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

// Sets the trailing slash behavior to pass to the `mux` layer. When
// `true`, routes with and without trailing slashes resolve to the
// same target. When `false`, the trailing slash is meaningful. The
// default value for Gimlet apps is `true`, and this method should be
// replaced with something more reasonable in the future.
func (self *ApiApp) SetStrictSlash(v bool) {
	self.strictSlash = v
}

// Run configured API service on the configured port. Registers
// middlewear for gziped responses and graceful shutdown with a 10
// second timeout.
func (self *ApiApp) Run() error {
	if self.isResolved == false {
		self.Resolve()
	}

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(NewAppLogger())
	n.Use(gzip.Gzip(gzip.DefaultCompression))

	n.UseHandler(self.router)

	listenOn := ":" + strconv.Itoa(self.port)
	grip.Noticeln("starting app on:", listenOn)

	graceful.Run(listenOn, 10*time.Second, n)
	return nil
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

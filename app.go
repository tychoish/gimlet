package gimlet

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/tychoish/grip"
)

//go:generate stringer -type=httpMethod
type httpMethod int

const (
	GET httpMethod = iota
	PUT
	POST
	DELETE
	PATCH
)

// Chainable method to add a handler for the GET method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Get() *ApiRoute {
	self.methods = append(self.methods, GET)
	return self
}

// Chainable method to add a handler for the PUT method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Put() *ApiRoute {
	self.methods = append(self.methods, PUT)
	return self
}

// Chainable method to add a handler for the POST method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Post() *ApiRoute {
	self.methods = append(self.methods, POST)
	return self
}

// Chainable method to add a handler for the DELETE method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Delete() *ApiRoute {
	self.methods = append(self.methods, DELETE)
	return self
}

// Chainable method to add a handler for the PATCH method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Patch() *ApiRoute {
	self.methods = append(self.methods, PATCH)
	return self
}

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

// Run configured API service on the configured port.
func (self *ApiApp) Run() error {
	if self.isResolved == false {
		self.Resolve()
	}

	n := negroni.New(negroni.NewRecovery(), newAppLogger())
	n.UseHandler(self.router)

	listenOn := ":" + strconv.Itoa(self.port)
	grip.Noticeln("starting app on:", listenOn)
	return http.ListenAndServe(listenOn, n)
}

// Represents each route in the application and includes the route and
// associate internal metadata for the route.
type ApiRoute struct {
	route   string
	methods []httpMethod
	handler http.HandlerFunc
	version int
}

// Checks if a route has is valid. Current implementation only makes
// sure that the version of the route is method.
func (self *ApiRoute) IsValid() bool {
	if self.version >= 0 {
		return true
	} else {
		return false
	}
}

// Specify an integer for the version of this route.
func (self *ApiRoute) Version(version int) *ApiRoute {
	if version < 0 {
		grip.Warningf("%d is not a valid version", version)
	} else {
		self.version = version
	}
	return self
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

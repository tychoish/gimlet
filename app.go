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

// Convience functions for setting method via method chaning
func (self *apiRoute) Get() *apiRoute {
	self.methods = append(self.methods, GET)
	return self
}
func (self *apiRoute) Put() *apiRoute {
	self.methods = append(self.methods, PUT)
	return self
}
func (self *apiRoute) Post() *apiRoute {
	self.methods = append(self.methods, POST)
	return self
}
func (self *apiRoute) Delete() *apiRoute {
	self.methods = append(self.methods, DELETE)
	return self
}
func (self *apiRoute) Patch() *apiRoute {
	self.methods = append(self.methods, PATCH)
	return self
}

type ApiApp struct {
	routes         []*apiRoute
	defaultVersion int
	isResolved     bool
	router         *mux.Router
	port           int
	strictSlash    bool
}

func (self *apiRoute) IsValid() bool {
	if self.version >= 0 {
		return true
	} else {
		return false
	}
}

// specify either an integer or api version constant. Versions are
// validated in the Resolve() method. Invalid input is ignored
func (self *apiRoute) Version(version int) *apiRoute {
	if version < 0 {
		grip.Warningf("%d is not a valid version", version)
	} else {
		self.version = version
	}
	return self
}

func (self *ApiApp) SetDefaultVersion(version int) {
	if version < 0 {
		grip.Warningf("%d is not a valid version", version)
	} else {
		self.defaultVersion = version
		grip.Noticef("Set default api version to /v%d/", version)
	}

}

type apiRoute struct {
	route   string
	methods []httpMethod
	handler http.HandlerFunc
	version int
}

func NewApp() *ApiApp {
	return &ApiApp{
		defaultVersion: 0, // this is the same as having no version prepended to the path.
		port:           8889,
		strictSlash:    true,
	}
}

func (self *ApiApp) SetPort(port int) (err error) {
	defaultPort := 3000

	if port == self.port {
		grip.Warningf("port is already set to %d", self.port)
		return
	} else if port <= 0 && self.port != defaultPort {
		err = fmt.Errorf("%d is not a valid port number, using %d", port, defaultPort)
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

// Sets the trailing slash behavior to pass to the `mux` layer. When
// `true`, routes with and without trailing slashes resolve to the
// same target. When `false`, the trailing slash is meaningful. The
// default value for Gimlet apps is `true`, and this method should be
// replaced with something more reasonable in the future.
func (self *ApiApp) SetStrictSlash(v bool) {
	self.strictSlash = v
}

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

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
	defaultVersion apiVersion
	isResolved     bool
	router         *mux.Router
	port           int
}

//go:generate stringer -type=apiVersion
type apiVersion int

const (
	invalidVersion            = -1
	None           apiVersion = iota
	V1
	V2
	V3
	V4
	V5
	V6
	V7
	V8
	V9
	V10
	v11
	v12
)

func (self *apiRoute) IsValid() bool {
	if self.version > 12 || self.version < 0 {
		return false
	} else {
		return true
	}
}

// specify either an integer or api version constant. Versions are
// validated in the Resolve() method. Invalid input is ignored
func (self *apiRoute) Version(v interface{}) *apiRoute {
	switch v := v.(type) {
	case int:
		self.version = apiVersion(v)
	case apiVersion:
		self.version = v
	}

	return self
}

func (self *ApiApp) SetDefaultVersion(v interface{}) {
	switch v := v.(type) {
	case int:
		self.defaultVersion = apiVersion(v)
	case apiVersion:
		self.defaultVersion = v
	}
}

type apiRoute struct {
	route   string
	methods []httpMethod
	handler http.HandlerFunc
	version apiVersion
}

func NewApp() *ApiApp {
	return &ApiApp{
		defaultVersion: 0, // this is the same as having no version prepended to the path.
		port:           3000,
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

func (self *ApiApp) Run() error {
	if self.isResolved == false {
		self.Resolve()
	}

	n := negroni.New(negroni.NewRecovery(), newAppLogger())
	n.UseHandler(self.router)

	listenOn := ":" + strconv.Itoa(self.port)
	return http.ListenAndServe(listenOn, n)
}

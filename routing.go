package gimlet

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tychoish/grip"
)

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

// Primary method for creating and registering a new route with an
// application. Use as the root of a method chain, passing this method
// the path of the route.
func (self *ApiApp) AddRoute(r string) *ApiRoute {
	route := &ApiRoute{route: r, version: -1}

	// data validation and cleanup
	if !strings.HasPrefix(route.route, "/") {
		route.route = "/" + route.route
	}

	self.routes = append(self.routes, route)

	return route
}

// Processes the data in an application and creats a mux.Router object.
func (self *ApiApp) Resolve() *mux.Router {
	router := mux.NewRouter()
	self.router = router

	for _, route := range self.routes {
		if !route.IsValid() {
			grip.Errorf("%d is an invalid api version. not adding route for %s",
				route.version, route.route)
			continue
		}

		var methods []string
		for _, m := range route.methods {
			methods = append(methods, strings.ToLower(m.String()))
		}

		if route.version > 0 {
			versionedRoute := fmt.Sprintf("/v%d%s", route.version, route.route)
			router.HandleFunc(versionedRoute, route.handler).Methods(methods...)
			grip.Debugln("added route for:", versionedRoute)

			if self.strictSlash {
				if strings.HasSuffix(versionedRoute, "/") {
					versionedRoute = strings.TrimRight(versionedRoute, "/")
				} else {
					versionedRoute = versionedRoute + "/"
				}
				router.HandleFunc(versionedRoute, route.handler).Methods(methods...)
			}
		}

		if route.version == self.defaultVersion || route.version == 0 {
			router.HandleFunc(route.route, route.handler).Methods(methods...)
			grip.Debugln("added route for:", route.route)

			if self.strictSlash {
				var newRoute string
				if strings.HasSuffix(route.route, "/") {
					newRoute = strings.TrimRight(route.route, "/")
				} else {
					newRoute = route.route + "/"
				}
				router.HandleFunc(newRoute, route.handler).Methods(methods...)
			}

		}
	}

	self.isResolved = true

	return router
}

// Processes an http.Request and returns a map of strings to decoded
// strings for all arguments passed to the method in the URL. Use this
// helper function when writing handler functions.
func GetVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

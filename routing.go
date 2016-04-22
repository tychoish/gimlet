package gimlet

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tychoish/grip"
)

// Represents each route in the application and includes the route and
// associate internal metadata for the route.
type APIRoute struct {
	route   string
	methods []httpMethod
	handler http.HandlerFunc
	version int
}

// Checks if a route has is valid. Current implementation only makes
// sure that the version of the route is method.
func (r *APIRoute) IsValid() bool {
	return r.version >= 0
}

// Specify an integer for the version of this route.
func (r *APIRoute) Version(version int) *APIRoute {
	if version < 0 {
		grip.Warningf("%d is not a valid version", version)
	} else {
		r.version = version
	}
	return r
}

// Primary method for creating and registering a new route with an
// application. Use as the root of a method chain, passing this method
// the path of the route.
func (a *APIApp) AddRoute(r string) *APIRoute {
	route := &APIRoute{route: r, version: -1}

	// data validation and cleanup
	if !strings.HasPrefix(route.route, "/") {
		route.route = "/" + route.route
	}

	a.routes = append(a.routes, route)

	return route
}

// Processes the data in an application and creats a mux.Router object.
func (a *APIApp) Resolve() error {
	a.router = mux.NewRouter().StrictSlash(a.strictSlash)

	var hasErrs bool
	for _, route := range a.routes {
		if !route.IsValid() {
			hasErrs = true
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
			a.router.HandleFunc(versionedRoute, route.handler).Methods(methods...)
			grip.Debugln("added route for:", versionedRoute)
		}

		if route.version == a.defaultVersion || route.version == 0 {
			a.router.HandleFunc(route.route, route.handler).Methods(methods...)
			grip.Debugln("added route for:", route.route)

		}
	}

	a.isResolved = true

	if hasErrs {
		return errors.New("encountered errors resolving routes")
	}

	return nil
}

// Processes an http.Request and returns a map of strings to decoded
// strings for all arguments passed to the method in the URL. Use this
// helper function when writing handler functions.
func GetVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

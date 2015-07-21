package gimlet

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tychoish/grip"
)

func (self *ApiApp) AddRoute(r string) *apiRoute {
	route := &apiRoute{route: r, version: invalidVersion}

	// data validation and cleanup
	if !strings.HasPrefix(route.route, "/") {
		route.route = "/" + route.route
	}

	self.routes = append(self.routes, route)

	return route
}

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

		if route.version != None {
			versionedRoute := fmt.Sprint("/", strings.ToLower(route.version.String()), route.route)

			router.HandleFunc(versionedRoute, route.handler).Methods(methods...)
			grip.Debugf("added route for:", versionedRoute)
		}

		if route.version == self.defaultVersion || route.version == None {
			router.HandleFunc(route.route, route.handler).Methods(methods...)
			grip.Debugln("added route for:", route.route)
		}
	}

	self.isResolved = true

	return router
}

func GetVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

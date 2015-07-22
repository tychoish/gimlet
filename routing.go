package gimlet

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tychoish/grip"
)

func (self *ApiApp) AddRoute(r string) *ApiRoute {
	route := &ApiRoute{route: r, version: -1}

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

func GetVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

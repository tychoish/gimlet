package gimlet

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tychoish/grip"
	"github.com/tychoish/grip/level"
	"github.com/tychoish/grip/send"
)

// AppSuite contains tests of the APIApp system. Tests of the route
// methods are ostly handled in other suites.
type AppSuite struct {
	constructor func() *APIApp
	app         *APIApp
	suite.Suite
}

func TestChiAppSuite(t *testing.T) {
	s := &AppSuite{}
	s.constructor = func() *APIApp { return NewApp().SetRouter(RouterImplChi) }
	suite.Run(t, s)
}

func TestGorilaAppSuite(t *testing.T) {
	s := &AppSuite{}
	s.constructor = func() *APIApp { return NewApp().SetRouter(RouterImplGorilla) }
	suite.Run(t, s)
}

func TestDefaultAppSuite(t *testing.T) {
	s := &AppSuite{}
	s.constructor = NewApp
	suite.Run(t, s)
}

func (s *AppSuite) SetupTest() {
	s.app = s.constructor()
	s.app.AddMiddleware(MakeRecoveryLogger())
	err := grip.Sender().SetLevel(send.LevelInfo{Default: level.Debug, Threshold: level.Info})
	s.NoError(err)
}

func (s *AppSuite) TestDefaultValuesAreSet() {
	s.app = NewApp()
	s.Len(s.app.middleware, 0)
	s.Len(s.app.routes, 0)
	s.Equal(s.app.port, 3000)
	s.True(s.app.StrictSlash)
	s.False(s.app.isResolved)
}

func (s *AppSuite) TestRouterGetterReturnsErrorWhenUnresovled() {
	s.False(s.app.isResolved)

	_, err := s.app.Router()
	s.Error(err)
}

func (s *AppSuite) TestMiddleWearResetEmptiesList() {
	s.app.AddMiddleware(NewAppLogger())
	s.app.AddMiddleware(NewStatic("", http.Dir("")))
	s.Len(s.app.middleware, 3)
	s.app.ResetMiddleware()
	s.Len(s.app.middleware, 0)
}

func (s *AppSuite) TestPortSetterDoesNotAllowImpermisableValues() {
	s.Equal(s.app.port, 3000)

	for _, port := range []int{0, -1, -2000, 99999, 65536, 1000, 100, 1023} {
		err := s.app.SetPort(port)
		s.Equal(s.app.port, 3000)
		s.Error(err)
	}

	for _, port := range []int{1025, 65535, 50543, 8080, 8000} {
		err := s.app.SetPort(port)
		s.Equal(s.app.port, port)
		s.NoError(err)
	}
}

func (s *AppSuite) TestAlternateMiddlewares() {
	s.Len(s.app.middleware, 1)
	s.app.AddMiddlewareHandler(MiddlewareFunc(NewAppLogger()))
	s.app.AddMiddlewareFunc(func(next http.HandlerFunc) http.HandlerFunc { return next })
	s.Len(s.app.middleware, 3)

	err := s.app.Resolve()
	s.NoError(err)
}

func (s *AppSuite) TestRouterReturnsRouterInstanceWhenResolved() {
	s.False(s.app.isResolved)

	switch s.app.routerImpl {
	case RouterImplChi:
		r, err := s.app.Mux()
		s.Nil(r)
		s.Error(err)
	case RouterImplGorilla:
		r, err := s.app.Router()
		s.Nil(r)
		s.Error(err)
	default:
		s.T().Fatal("unsupported router")
	}

	s.app.AddRoute("/foo").Version(1).Get().Handler(func(_ http.ResponseWriter, _ *http.Request) {})
	s.NoError(s.app.Resolve())
	s.True(s.app.isResolved)

	switch s.app.routerImpl {
	case RouterImplChi:
		r, err := s.app.Mux()
		s.NotNil(r)
		s.NoError(err)
	case RouterImplGorilla:
		r, err := s.app.Router()
		s.NotNil(r)
		s.NoError(err)
	default:
		s.T().Fatal("unsupported router")
	}
}

func (s *AppSuite) TestResolveEncountersErrorsWithAnInvalidRoot() {
	s.False(s.app.isResolved)

	s.app.AddRoute("/foo").Version(-10)
	err1 := s.app.Resolve()
	s.Error(err1)

	// also check that app.getNegroni
	n, err2 := s.app.getNegroni()
	s.Nil(n)
	s.Error(err2)

	// also to run
	err2 = s.app.Run(context.TODO())
	s.Error(err2)
}

func (s *AppSuite) TestSetPortToExistingValueIsANoOp() {
	port := s.app.port

	s.Equal(port, s.app.port)
	s.NoError(s.app.SetPort(port))
	s.Equal(port, s.app.port)
}

func (s *AppSuite) TestResolveValidRoute() {
	s.False(s.app.isResolved)
	route := &APIRoute{
		version: 1,
		methods: []httpMethod{get},
		handler: func(_ http.ResponseWriter, _ *http.Request) { grip.Info("hello") },
		route:   "/foo",
	}
	s.True(route.IsValid())
	s.app.routes = append(s.app.routes, route)
	s.NoError(s.app.Resolve())
	s.True(s.app.isResolved)
	n, err := s.app.getNegroni()
	s.NotNil(n)
	s.NoError(err)
}

func (s *AppSuite) TestResolveAppWithDefaultVersion() {
	s.app.NoVersions = true
	s.False(s.app.isResolved)
	route := &APIRoute{
		version: -1,
		methods: []httpMethod{get},
		handler: func(_ http.ResponseWriter, _ *http.Request) { grip.Info("hello") },
		route:   "/foo",
	}
	s.True(route.IsValid())
	s.app.routes = append(s.app.routes, route)
	s.NoError(s.app.Resolve())
	s.True(s.app.isResolved)
}

func (s *AppSuite) TestResolveAppWithInvaldVersion() {
	s.app.NoVersions = false
	s.False(s.app.isResolved)
	route := &APIRoute{
		version: -1,
		methods: []httpMethod{get},
		handler: func(_ http.ResponseWriter, _ *http.Request) { grip.Info("hello") },
		route:   "/foo",
	}
	s.True(route.IsValid())
	s.app.routes = append(s.app.routes, route)
	s.Error(s.app.Resolve())
	s.False(s.app.isResolved)
}

func (s *AppSuite) TestSetHostOperations() {
	s.Equal("", s.app.address)
	s.False(s.app.isResolved)

	s.NoError(s.app.SetHost("1"))
	s.Equal("1", s.app.address)
	s.app.isResolved = true

	s.Error(s.app.SetHost("2"))
	s.Equal("1", s.app.address)
}

func (s *AppSuite) TestSetPrefix() {
	s.Equal("", s.app.prefix)

	s.app.SetPrefix("foo")
	s.Equal("/foo", s.app.prefix)
	s.app.SetPrefix("/bar")
	s.Equal("/bar", s.app.prefix)
}

func (s *AppSuite) TestHandlerGetter() {
	s.NoError(s.app.Resolve())
	hone, err := s.app.getNegroni()
	s.NoError(err)
	s.NotNil(hone)
	htwo, err := s.app.Handler()
	s.NoError(err)
	s.NotNil(htwo)

	// should be equivalent results but are different instances, as each app should be distinct.
	s.NotEqual(hone, htwo)
}

func (s *AppSuite) TestAppRun() {
	s.Len(s.app.routes, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.NoError(s.app.Resolve())
	s.NoError(s.app.Run(ctx))
}

func (s *AppSuite) TestAppRunWithError() {
	s.Len(s.app.routes, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.app.port = -10
	s.app.address = ":;;;:::::"
	wait, err := s.app.BackgroundRun(ctx)
	s.Error(err)
	s.Nil(wait)
}

func (s *AppSuite) TestWrapperAccessors() {
	s.Len(s.app.wrappers, 0)
	s.app.AddWrapper(MakeRecoveryLogger())
	s.app.AddWrapperHandler(MiddlewareFunc(MakeRecoveryLogger()))
	s.app.AddWrapperFunc(func(next http.HandlerFunc) http.HandlerFunc { return next })

	s.Len(s.app.wrappers, 3)
	s.app.RestWrappers()
	s.Len(s.app.wrappers, 0)
}

func (s *AppSuite) TestResolveWithInvalidMiddleware() {
	s.app.middleware = append(s.app.middleware, 1, true, "wat")
	s.Error(s.app.Resolve())
}

func (s *AppSuite) TestResolveWithInvalidWrappers() {
	s.app.wrappers = append(s.app.wrappers, 1, true, "wat")
	s.Error(s.app.Resolve())

}

func (s *AppSuite) TestResolveWithInvalidRouteWrappers() {
	r := s.app.AddRoute("/what").Version(1).Get().Handler(func(_ http.ResponseWriter, _ *http.Request) {})
	r.wrappers = append(r.wrappers, 1, true, "wat")
	s.Error(s.app.Resolve())
}

func (s *AppSuite) TestResolveWithRouteWrappers() {
	s.app.AddRoute("/what").Version(1).Get().Handler(func(_ http.ResponseWriter, _ *http.Request) {}).
		Wrap(MakeRecoveryLogger()).
		WrapHandler(func(next http.Handler) http.Handler { return next }).
		WrapHandler(MiddlewareFunc(MakeRecoveryLogger())).
		WrapHandlerFunc(func(next http.HandlerFunc) http.HandlerFunc { return next })
	s.NoError(s.app.Resolve())
}

func (s *AppSuite) TestResolveAndHandlerWithInvalidRoutes() {
	s.app = &APIApp{}
	s.Error(s.app.routerImpl.Validate())

	s.Error(s.app.Resolve())
	h, err := s.app.Handler()
	s.Nil(h)
	s.Error(err)
}

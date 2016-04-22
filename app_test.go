package gimlet

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestGoCheckTests(t *testing.T) { TestingT(t) }

type GimletSuite struct {
	app *APIApp
}

var _ = Suite(&GimletSuite{})

func (s *GimletSuite) SetUpTest(c *C) {
	s.app = NewApp()
}

func (s *GimletSuite) TestDefaultValuesAreSet(c *C) {
	c.Assert(s.app.middleware, HasLen, 3)
	c.Assert(s.app.routes, HasLen, 0)
	c.Assert(s.app.port, Equals, 3000)
	c.Assert(s.app.strictSlash, Equals, true)
	c.Assert(s.app.isResolved, Equals, false)
	c.Assert(s.app.defaultVersion, Equals, -1)
}

func (s *GimletSuite) TestRouterGetterReturnsErrorWhenUnresovled(c *C) {
	c.Assert(s.app.isResolved, Equals, false)

	_, err := s.app.Router()
	c.Assert(err, Not(IsNil))
}

func (s *GimletSuite) TestDefaultVersionSetter(c *C) {
	c.Assert(s.app.defaultVersion, Equals, -1)
	s.app.SetDefaultVersion(-2)
	c.Assert(s.app.defaultVersion, Equals, -1)

	s.app.SetDefaultVersion(0)
	c.Assert(s.app.defaultVersion, Equals, 0)

	s.app.SetDefaultVersion(1)
	c.Assert(s.app.defaultVersion, Equals, 1)

	for idx := range [100]int{} {
		s.app.SetDefaultVersion(idx)
		c.Assert(s.app.defaultVersion, Equals, idx)
	}
}

func (s *GimletSuite) TestMiddleWearResetEmptiesList(c *C) {
	c.Assert(s.app.middleware, HasLen, 3)
	s.app.ResetMiddleware()
	c.Assert(s.app.middleware, HasLen, 0)
}

func (s *GimletSuite) TestMiddleWearAdderAddsItemToList(c *C) {
	c.Assert(s.app.middleware, HasLen, 3)
	s.app.AddMiddleware(NewAppLogger())
	c.Assert(s.app.middleware, HasLen, 4)
}

func (s *GimletSuite) TestPortSetterDoesNotAllowImpermisableValues(c *C) {
	c.Assert(s.app.port, Equals, 3000)

	for _, port := range []int{0, -1, -2000, 99999, 65536, 1000, 100, 1023} {
		err := s.app.SetPort(port)
		c.Assert(s.app.port, Equals, 3000)
		c.Assert(err, Not(IsNil))
	}

	for _, port := range []int{1025, 65535, 50543, 8080, 8000} {
		err := s.app.SetPort(port)
		c.Assert(s.app.port, Equals, port)
		c.Assert(err, IsNil)
	}
}

func (s *GimletSuite) TestAddAppReturnsErrorIfOuterAppIsResolved(c *C) {
	newApp := NewApp()
	err := newApp.Resolve()
	c.Assert(err, IsNil)
	c.Assert(newApp.isResolved, Equals, true)

	// if you attempt use AddApp on an app that is already
	// resolved, it returns an error.
	c.Assert(newApp.AddApp(s.app), Not(IsNil))
}

func (s *GimletSuite) TestRouteMergingInIfVersionsAreTheSame(c *C) {
	subApp := NewApp()
	c.Assert(subApp.routes, HasLen, 0)
	route := subApp.AddRoute("/foo")
	c.Assert(subApp.routes, HasLen, 1)

	c.Assert(s.app.routes, HasLen, 0)
	err := s.app.AddApp(subApp)
	c.Assert(err, IsNil)

	c.Assert(s.app.routes, HasLen, 1)
	c.Assert(s.app.routes[0], Equals, route)
}

func (s *GimletSuite) TestRouteMergingInWithDifferntVersions(c *C) {
	// If the you have two apps with different default versions,
	// routes in the sub-app that don't have a version set, should
	// get their version set to whatever the value of the sub
	// app's default value at the time of merging the apps.
	subApp := NewApp()
	subApp.SetDefaultVersion(2)
	c.Assert(s.app.defaultVersion, Not(Equals), subApp.defaultVersion)

	// add a route to the first app
	c.Assert(subApp.routes, HasLen, 0)
	route := subApp.AddRoute("/foo").Version(3)
	c.Assert(route.version, Equals, 3)
	c.Assert(subApp.routes, HasLen, 1)

	// try adding to second app, to the first, with one route
	c.Assert(s.app.routes, HasLen, 0)
	err := s.app.AddApp(subApp)
	c.Assert(err, IsNil)
	c.Assert(s.app.routes, HasLen, 1)
	c.Assert(s.app.routes[0], Equals, route)

	nextApp := NewApp()
	c.Assert(nextApp.routes, HasLen, 0)
	nextRoute := nextApp.AddRoute("/bar")
	c.Assert(nextApp.routes, HasLen, 1)
	c.Assert(nextRoute.version, Equals, -1)
	nextApp.SetDefaultVersion(3)
	c.Assert(nextRoute.version, Equals, -1)

	// make sure the default value of nextApp is on the route in the subApp
	err = s.app.AddApp(nextApp)
	c.Assert(err, IsNil)
	c.Assert(s.app.routes[1], Equals, nextRoute)

	// this is the meaningful validation here.
	c.Assert(s.app.routes[1].version, Equals, 3)
}

package gimlet

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeApplications(t *testing.T) {
	for name, makeApp := range map[string]func() *APIApp{
		"Default": NewApp,
		"Gorilla": func() *APIApp {
			return NewApp().SetRouter(RouterImplGorilla)
		},
		"Chi": func() *APIApp {
			return NewApp().SetRouter(RouterImplChi)
		},
		"GorillaNoSlash": func() *APIApp {
			app := NewApp().SetRouter(RouterImplGorilla)
			app.StrictSlash = false
			return app

		},
		"ChiNoSlash": func() *APIApp {
			app := NewApp().SetRouter(RouterImplChi)
			app.StrictSlash = false
			return app
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Run("MergeAppsIntoRoute", func(t *testing.T) {
				// error when no apps
				h, err := MergeApplications()
				assert.Error(t, err)
				assert.Nil(t, h)

				// one app should merge just fine
				app := makeApp()
				app.SetPrefix("foo")
				app.AddMiddleware(MakeRecoveryLogger())
				app.AddMiddlewareHandler(MiddlewareFunc(MakeRecoveryLogger()))
				app.AddMiddlewareFunc(func(next http.HandlerFunc) http.HandlerFunc { return next })

				h, err = MergeApplications(app)
				assert.NoError(t, err)
				assert.NotNil(t, h)

				// a bad app should error
				bad := makeApp()
				bad.AddRoute("/foo").version = -1

				h, err = MergeApplications(bad)
				assert.Error(t, err)
				assert.Nil(t, h)

				// even when it's combined with a good one
				h, err = MergeApplications(bad, app)
				assert.Error(t, err)
				assert.Nil(t, h)
			})
			t.Run("Handler", func(t *testing.T) {
				// one app should merge just fine
				app := makeApp()
				app.SetPrefix("foo")
				app.AddMiddleware(MakeRecoveryLogger())
				app.AddMiddlewareHandler(MiddlewareFunc(MakeRecoveryLogger()))
				app.AddMiddlewareFunc(func(next http.HandlerFunc) http.HandlerFunc { return next })
				h, err := app.Handler()
				assert.NoError(t, err)
				assert.NotNil(t, h)
			})
			t.Run("Resolve", func(t *testing.T) {
				// one app should merge just fine
				app := makeApp()
				app.SetPrefix("foo")
				app.AddMiddleware(MakeRecoveryLogger())
				app.AddMiddlewareHandler(MiddlewareFunc(MakeRecoveryLogger()))
				app.AddMiddlewareFunc(func(next http.HandlerFunc) http.HandlerFunc { return next })
				assert.NoError(t, app.Resolve())
			})
			t.Run("MergeApps", func(t *testing.T) {
				// one app should merge just fine
				app := makeApp()
				app.AddMiddleware(MakeRecoveryLogger())
				app.AddMiddlewareHandler(MiddlewareFunc(MakeRecoveryLogger()))

				err := app.Merge(app)
				assert.NoError(t, err)

				// can't merge more than once
				err = app.Merge(app)
				assert.Error(t, err)
				app.hasMerged = false

				err = app.Merge()
				assert.Error(t, err)
				app.hasMerged = false

				// check duplicate without prefix or middleware
				app.AddRoute("/foo").Version(2).Get()
				bad := makeApp()
				bad.AddRoute("/foo").Version(2).Get()
				err = app.Merge(bad)
				assert.Error(t, err)
				app.hasMerged = false

				app.SetPrefix("foo")
				assert.Error(t, app.Merge(bad))
				app.prefix = ""

				app2 := makeApp()
				app2.SetPrefix("foo")
				app2.AddMiddleware(MakeRecoveryLogger())
				app2.AddMiddlewareHandler(MiddlewareFunc(MakeRecoveryLogger()))

				err = app.Merge(app2)
				assert.NoError(t, err)
				app.hasMerged = false

				err = app.Merge(app2, app2)
				assert.Error(t, err)
				app.hasMerged = false

				// can't have duplicated methods
				app2 = makeApp()
				app2.SetPrefix("foo")
				app2.AddRoute("/foo").Version(2).Get()
				err = app.Merge(app2)
				assert.NoError(t, err)
				app.hasMerged = false

				app3 := makeApp()
				app3.AddMiddleware(MakeRecoveryLogger())
				app3.AddMiddlewareHandler(MiddlewareFunc(MakeRecoveryLogger()))
				app3.SetPrefix("/wat")
				app3.AddRoute("/bar").Version(3).Get()
				err = app.Merge(app3)
				assert.NoError(t, err)
				app.hasMerged = false

				app2.AddMiddleware(MakeRecoveryLogger())
				app3.prefix = ""
				err = app.Merge(app2, app3)
				assert.Error(t, err)
				app.hasMerged = false

			})
		})
	}
}

func TestAssembleHandler(t *testing.T) {
	t.Run("Gorilla", func(t *testing.T) {
		router := mux.NewRouter()

		app := NewApp()
		app.AddRoute("/foo").version = -1

		h, err := AssembleHandlerGorilla(router, app)
		assert.Error(t, err)
		assert.Nil(t, h)

		app = NewApp()
		app.SetPrefix("foo")
		app.AddMiddleware(MakeRecoveryLogger())

		h, err = AssembleHandlerGorilla(router, app)
		assert.NoError(t, err)
		assert.NotNil(t, h)

		// can't have duplicated methods
		app2 := NewApp()
		app2.SetPrefix("foo")
		h, err = AssembleHandlerGorilla(router, app, app2)
		assert.Error(t, err)
		assert.Nil(t, h)

		app = NewApp()
		app.StrictSlash = false
		app.AddMiddleware(MakeRecoveryLogger())
		h, err = AssembleHandlerGorilla(router, app)
		assert.NoError(t, err)
		assert.NotNil(t, h)
	})
	t.Run("Chi", func(t *testing.T) {
		app := NewApp()
		app.AddRoute("/foo").version = -1

		h, err := AssembleHandlerChi(chi.NewMux(), app)
		assert.Error(t, err)
		assert.Nil(t, h)

		app = NewApp()
		app.SetPrefix("foo")
		app.AddMiddleware(MakeRecoveryLogger())

		h, err = AssembleHandlerChi(chi.NewMux(), app)
		assert.NoError(t, err)
		assert.NotNil(t, h)

		// can't have duplicated methods
		app2 := NewApp()
		app2.SetPrefix("foo")
		h, err = AssembleHandlerChi(chi.NewMux(), app, app2)
		assert.Error(t, err)
		assert.Nil(t, h)

		app = NewApp()
		app.StrictSlash = false
		app.AddMiddleware(MakeRecoveryLogger())
		h, err = AssembleHandlerChi(chi.NewMux(), app)
		assert.NoError(t, err)
		assert.NotNil(t, h)

	})
	t.Run("MixedRouters", func(t *testing.T) {
		_, err := MergeApplications(NewApp(), NewApp())
		assert.NoError(t, err)
		_, err = MergeApplications(&APIApp{}, &APIApp{})
		assert.Error(t, err)
		_, err = MergeApplications(NewApp().SetRouter(RouterImplChi), NewApp().SetRouter(RouterImplGorilla))
		assert.Error(t, err)
	})
	t.Run("InvalidRouterType", func(t *testing.T) {
		// this is really internal, but it's nice to be safe
		app := NewApp()
		app.AddRoute("/foo").Version(1).Get().Handler(func(_ http.ResponseWriter, _ *http.Request) {})

		require.Error(t, app.attachRoutes(true, true))
	})

}

func TestAppContainsRoute(t *testing.T) {
	app := NewApp()
	app.AddRoute("/foo").Get().Version(2)

	assert.False(t, app.containsRoute("/foo", 1, []httpMethod{patch}))
	assert.False(t, app.containsRoute("/foo", 2, []httpMethod{patch}))
	assert.False(t, app.containsRoute("/bar", 1, []httpMethod{get}))
	assert.False(t, app.containsRoute("/bar", 2, []httpMethod{get}))
	assert.False(t, app.containsRoute("/foo", 1, []httpMethod{get}))
	assert.True(t, app.containsRoute("/foo", 2, []httpMethod{get}))
}

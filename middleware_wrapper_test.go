package gimlet

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMiddleware(t *testing.T) {
	t.Run("ValidatesInput", func(t *testing.T) {
		assert.Panics(t, func() {
			MergeMiddleware()
		})
	})
	t.Run("ValidateCalling", func(t *testing.T) {
		legacyCalls := 0
		legacyFunc := func(h http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				legacyCalls++

				h(w, r)
			}
		}

		noop := func(w http.ResponseWriter, r *http.Request) {}

		MergeMiddleware(WrapperMiddleware(legacyFunc), WrapperMiddleware(legacyFunc)).ServeHTTP(nil, nil, noop)
		assert.Equal(t, 2, legacyCalls)
	})

}

func TestMiddlewareFuncWrapper(t *testing.T) {
	legacyCalls := 0
	legacyFunc := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			legacyCalls++

			h(w, r)
		}
	}

	nextCalls := 0
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
	}

	wrapped := WrapperMiddleware(legacyFunc)
	assert.Implements(t, (*Middleware)(nil), wrapped)
	assert.Equal(t, 0, legacyCalls)
	assert.Equal(t, 0, nextCalls)

	wrapped.ServeHTTP(nil, nil, next)
	assert.Equal(t, 1, legacyCalls)
	assert.Equal(t, 1, nextCalls)
}

func TestMiddlewareWrapper(t *testing.T) {
	legacyCalls := 0
	legacyFunc := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			legacyCalls++

			h.ServeHTTP(w, r)
		})
	}

	nextCalls := 0
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
	}

	wrapped := WrapperHandlerMiddleware(legacyFunc)
	assert.Implements(t, (*Middleware)(nil), wrapped)
	assert.Equal(t, 0, legacyCalls)
	assert.Equal(t, 0, nextCalls)

	wrapped.ServeHTTP(nil, nil, next)

	assert.Equal(t, 1, legacyCalls)
	assert.Equal(t, 1, nextCalls)
}

func TestMiddlewareFunction(t *testing.T) {
	legacyCalls := 0
	mw := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			legacyCalls++

			h.ServeHTTP(w, r)
		})
	}

	nextCalls := 0
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
	}

	wrapped := WrapperHandlerMiddleware(mw)

	mwfunc := MiddlewareFunc(wrapped)

	assert.Equal(t, 0, legacyCalls)
	assert.Equal(t, 0, nextCalls)

	mwfunc(http.HandlerFunc(next)).ServeHTTP(nil, nil)

	assert.Equal(t, 1, legacyCalls)
	assert.Equal(t, 1, nextCalls)
}

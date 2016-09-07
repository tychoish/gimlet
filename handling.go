package gimlet

import (
	"encoding/json"
	"net/http"

	"github.com/tychoish/grip"
)

// Handler makes it possible to register an http.HandlerFunc with a
// route. Chainable. The common pattern for implementing these
// functions is to write functions and methods in your application
// that *return* handler fucntions, so you can pass application state
// or other data into to the handlers when the applications start,
// without relying on either global state *or* running into complex
// typing issues.
func (m *APIRoute) Handler(h http.HandlerFunc) *APIRoute {
	m.handler = h

	return m
}

// WriteJSONResponse writes a JSON document to the body of an HTTP
// request, setting the return status of to 500 if the JSON
// seralization process encounters an error, otherwise return
func WriteJSONResponse(w http.ResponseWriter, code int, data interface{}) {
	j := &JSONMessage{data: data}

	out, err := j.MarshalPretty()

	if err != nil {
		grip.CatchDebug(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grip.Debug(j)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	size, err := w.Write(out)
	if err == nil {
		grip.Debugf("response object was %d", size)
	} else {
		grip.Warningf("encountered error %s writing a %d response", err.Error(), size)
	}
}

// WriteJSON is a helper method to write JSON data to the body of an
// HTTP request and return 200 (successful.)
func WriteJSON(w http.ResponseWriter, data interface{}) {
	// 200
	WriteJSONResponse(w, http.StatusOK, data)
}

// WriteErrorJSON is a helper method to write JSON data to the body of
// an HTTP request and return 400 (user error.)
func WriteErrorJSON(w http.ResponseWriter, data interface{}) {
	// 400
	WriteJSONResponse(w, http.StatusBadRequest, data)
}

// WriteInternalErrorJSON is a helper method to write JSON data to the
// body of an HTTP request and return 500 (internal error.)
func WriteInternalErrorJSON(w http.ResponseWriter, data interface{}) {
	// 500
	WriteJSONResponse(w, http.StatusInternalServerError, data)
}

// GetJSON parses JSON from a request body into an object specified by
// the request. Used in handler functiosn to retreve and parse data
// submitted by the client.
func GetJSON(r *http.Request, data interface{}) error {
	d := json.NewDecoder(r.Body)

	err := d.Decode(data)
	grip.CatchDebug(err)
	grip.Debug(&JSONMessage{data: data})

	return err
}

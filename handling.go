package gimlet

import (
	"encoding/json"
	"net/http"

	"github.com/tychoish/grip"
)

// JSONMessage is an implementation of the grip/message.Composer
// interface, used so that we can log all incoming and outgoing Json
// content in debug mode, without needing to serialize structs under
// normal operation. Also contains a MarshalPretty() method which is
// used in rendering JSON into the response objects.
type JSONMessage struct {
	data interface{}
}

func (m *JSONMessage) Resolve() string {
	out, err := json.Marshal(m.data)
	if err != nil {
		grip.CatchWarning(err)
		return ""
	}
	return string(out)
}

// In this implementation this is always true, but potentially, the
// message can force itm to be *not* logable. May be useful in the
// future to modify this form to suppress sensitive data.
func (m *JSONMessage) Loggable() bool {
	return true
}

// Return the data without seralizing it first. Useful for logging
// mechanisms that handle a raw format for insertion into a database
// or posting to a service.
func (m *JSONMessage) Raw() interface{} {
	return m.data
}

// A helper method to simplify calls to json.MarshalIndent(). This is
// not part of the Composer interface.
func (m *JSONMessage) MarshalPretty() ([]byte, error) {
	return json.MarshalIndent(m.data, "", "  ")
}

// Register an http.HandlerFunc with a route. Chainable. The common
// pattern for implementing these functions is to write functions and
// methods in your application that *return* handler fucntions, so you
// can pass application state or other data into to the handlers when
// the applications start, without relying on either global state *or*
// running into complex typing issues.
func (m *APIRoute) Handler(h http.HandlerFunc) *APIRoute {
	m.handler = h

	return m
}

// Writes a JSON document to the body of an HTTP request, setting the
// return status of to 500 if the JSON seralization process encounters
// an error, otherwise return
func WriteJSONResponse(w http.ResponseWriter, code int, data interface{}) {
	j := &JSONMessage{data: data}

	out, err := j.MarshalPretty()

	if err != nil {
		grip.CatchDebug(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grip.ComposeDebug(j)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(out)
	w.Write([]byte("\n"))
}

// A helper method to write JSON data to the body of an HTTP request and return 200 (successful.)
func WriteJSON(w http.ResponseWriter, data interface{}) {
	// 200
	WriteJSONResponse(w, http.StatusOK, data)
}

// A helper method to write JSON data to the body of an HTTP request and return 400 (user error.)
func WriteErrorJSON(w http.ResponseWriter, data interface{}) {
	// 400
	WriteJSONResponse(w, http.StatusBadRequest, data)
}

// A helper method to write JSON data to the body of an HTTP request and return 500 (internal error.)
func WriteInternalErrorJSON(w http.ResponseWriter, data interface{}) {
	// 500
	WriteJSONResponse(w, http.StatusInternalServerError, data)
}

// Parses JSON from a request body into an object specified by the
// request. Used in handler functiosn to retreve and parse data
// submitted by the client.
func GetJSON(r *http.Request, data interface{}) error {
	d := json.NewDecoder(r.Body)

	err := d.Decode(data)
	grip.CatchDebug(err)
	grip.ComposeDebug(&JSONMessage{data: data})

	return err
}

package gimlet

import (
	"encoding/json"
	"net/http"
)

// Register an http.HandlerFunc with a route. Chainable. The common
// pattern for implementing these functions is to write functions and
// methods in your application that *return* handler fucntions, so you
// can pass application state or other data into to the handlers when
// the applications start, without relying on either global state *or*
// running into complex typing issues.
func (self *ApiRoute) Handler(h http.HandlerFunc) *ApiRoute {
	self.handler = h

	return self
}

// Writes a JSON document to the body of an HTTP request, setting the
// return status of to 500 if the JSON seralization process encounters
// an error, otherwise return
func WriteJSONResponse(w http.ResponseWriter, code int, data interface{}) {
	out, err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	return d.Decode(data)
}

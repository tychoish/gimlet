package gimlet

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const maxRequestSize = 16 * 1024 * 1024 // 16 MB

// GetVars is a helper method that processes an http.Request and
// returns a map of strings to decoded strings for all arguments
// passed to the method in the URL. Use this helper function when
// writing handler functions.
//
// GetVars only works with the Gorilla Mux rotuer and not with the chi router.
func GetVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// GetParam provides a common interface for getting a URL parameter
// that uses gorilla/mux or chi.
func GetParam(r *http.Request, k string) string {
	if vars := GetVars(r); vars != nil {
		return vars[k]
	}

	return chi.URLParam(r, k)
}

// SetURLVars sets URL variables for testing purposes only.
func SetURLVars(r *http.Request, val map[string]string) *http.Request {
	return mux.SetURLVars(r, val)
}

// GetJSON parses JSON from a io.ReadCloser (e.g. http/*Request.Body
// or http/*Response.Body) into an object specified by the
// request. Used in handler functiosn to retreve and parse data
// submitted by the client.
//
// Returns an error if the body is greater than 16 megabytes in size.
func GetJSON(r io.ReadCloser, data interface{}) error {
	if r == nil {
		return errors.New("no data defined")
	}
	defer r.Close()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r, N: maxRequestSize})
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(json.Unmarshal(bytes, data))
}

// GetJSONUnlimited reads data from a io.ReadCloser, as with GetJSON,
// but does not bound the size of the request.
func GetJSONUnlimited(r io.ReadCloser, data interface{}) error {
	if r == nil {
		return errors.New("no data defined")
	}
	defer r.Close()

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(json.Unmarshal(bytes, data))
}

// GetYAML parses YAML from a io.ReadCloser (e.g. http/*Request.Body
// or http/*Response.Body) into an object specified by the
// request. Used in handler functiosn to retreve and parse data
// submitted by the client.u
func GetYAML(r io.ReadCloser, data interface{}) error {
	if r == nil {
		return errors.New("no data defined")
	}
	defer r.Close()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r, N: maxRequestSize})
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(yaml.Unmarshal(bytes, data))
}

// GetYAMLUnlimited reads data from a io.ReadCloser, as with GetYAML,
// but does not bound the size of the request.
func GetYAMLUnlimited(r io.ReadCloser, data interface{}) error {
	if r == nil {
		return errors.New("no data defined")
	}
	defer r.Close()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r, N: maxRequestSize})
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(yaml.Unmarshal(bytes, data))
}

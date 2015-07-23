package gimlet

import (
	"encoding/json"
	"net/http"
)

func (self *ApiRoute) Handler(h http.HandlerFunc) *ApiRoute {
	self.handler = h

	return self
}

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

func WriteJSON(w http.ResponseWriter, data interface{}) {
	// 200
	WriteJSONResponse(w, http.StatusOK, data)
}

func WriteErrorJSON(w http.ResponseWriter, data interface{}) {
	// 400
	WriteJSONResponse(w, http.StatusBadRequest, data)
}

func WriteInternalErrorJSON(w http.ResponseWriter, data interface{}) {
	// 500
	WriteJSONResponse(w, http.StatusInternalServerError, data)
}

func GetJSON(r *http.Request, data interface{}) error {
	d := json.NewDecoder(r.Body)
	return d.Decode(data)
}

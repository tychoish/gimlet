package gimlet

//go:generate stringer -type=httpMethod
type httpMethod int

// Typed constants for specifying HTTP method types on routes.
const (
	GET httpMethod = iota
	PUT
	POST
	DELETE
	PATCH
)

// Get is a chainable method to add a handler for the GET method to
// the current route. Routes may specify multiple methods.
func (r *APIRoute) Get() *APIRoute {
	r.methods = append(r.methods, GET)
	return r
}

// Put is a chainable method to add a handler for the PUT method to
// the current route. Routes may specify multiple methods.
func (r *APIRoute) Put() *APIRoute {
	r.methods = append(r.methods, PUT)
	return r
}

// Post is a chainable method to add a handler for the POST method to
// the current route. Routes may specify multiple methods.
func (r *APIRoute) Post() *APIRoute {
	r.methods = append(r.methods, POST)
	return r
}

// Delete is a chainable method to add a handler for the DELETE method
// to the current route. Routes may specify multiple methods.
func (r *APIRoute) Delete() *APIRoute {
	r.methods = append(r.methods, DELETE)
	return r
}

// Patch is a chainable method to add a handler for the PATCH method
// to the current route. Routes may specify multiple methods.
func (r *APIRoute) Patch() *APIRoute {
	r.methods = append(r.methods, PATCH)
	return r
}

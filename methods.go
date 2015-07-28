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

// Chainable method to add a handler for the GET method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Get() *ApiRoute {
	self.methods = append(self.methods, GET)
	return self
}

// Chainable method to add a handler for the PUT method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Put() *ApiRoute {
	self.methods = append(self.methods, PUT)
	return self
}

// Chainable method to add a handler for the POST method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Post() *ApiRoute {
	self.methods = append(self.methods, POST)
	return self
}

// Chainable method to add a handler for the DELETE method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Delete() *ApiRoute {
	self.methods = append(self.methods, DELETE)
	return self
}

// Chainable method to add a handler for the PATCH method to the
// current route. Routes may specify multiple methods.
func (self *ApiRoute) Patch() *ApiRoute {
	self.methods = append(self.methods, PATCH)
	return self
}

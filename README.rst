=====================================
''``gimlet`` -- HTTP Toolkit for APIs
=====================================

Overview
--------

``gimlet`` is a simple collection of tools for creating simple versioned
JSON/HTTP APIs. It builds on standard library tools and components of `gorilla
<http://www.gorillatoolkit.org/>`_ (mux) and `negroni
<https://github.com/urfave/negroni>`_, but these are not material to day to
say use.

See the `API documentation <https://godoc.org/github.com/deciduosity/gimlet>`_
for a complete overview.

Features
--------

- Authentication middleware and system, to permit integration with LDAP and
  Okta-based user providers. 
  
- Default middlewares for common operations including logging and
  compression.

- Methods for defining routes and applications with versioning, using
  idiomatic components and a clear ergonomic chainable API.

Goals
-----

- Allow developers to implement HTTP/JSON APIs by writing
  `http.HandlerFunc <http://golang.org/pkg/net/#HandlerFunc>`_ methods
  and passing `encoding/json <http://golang.org/pkg/encoding/json/>`_ 
  marshallable types to simple response-writing methods. 
  
- Make it easy to define a set of routes with a version prefix, and
  manage the version prefix at the routing layer rather than in
  handlers. 
  
- Reuse common components as necessary, and avoid recreating existing
  tools or making a large inflexible tool.

- Support common workflows, like authentication and observability.

Improve authentication workflows and integration. 
  
In short I was writing a JSON/HTTP API, and wanted the above
properties and found that I had written a little library that didn't
really have anything to do with the app I was writing so I'm spinning
it out both because *I* want to use this in my next project and I hope
you may find it useful.

You can `read an introductory blog post
<http://www.tychoish.com/posts/have-a-gimlet-a-go-json-http-api-toolkit/>`_,
if you like.

Development
-----------

Feel free (and encouraged!) to open issues or submit pull requests, if you
discover issues, or would like to use gimlet, but encounter an issue. 

Gimlet is available under the terms of the Apache License (v2).

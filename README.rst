===================================
``gimlet`` -- HTTP/JSON API Toolkit
===================================

``gimlet`` is a simple collection of tools for creating simple
versioned JSON/HTTP APIs. It builds on standard library tools and
components of `gorilla <>`_ (mux) and `negroni <>`_. 

The goal: 

- Allow developers to implement HTTP/JSON APIs by writing
  `http.HandlerFunc <>`_ methods and passing `encoding/json <>`_
  marshallable types to simple response-writing methods. 
  
- Make it easy to define a set of routes with a version prefix, and
  manage the version prefix at the routing layer rather than in
  handlers. 
  
- Reuse common components as necessary, and avoid recreating existing
  tools or making a large inflexible tool.
  
In short I was writing a JSON/HTTP API, and wanted the above
properties and found that I had written a little library that didn't
really have anything to do with the app I was writing so I'm spinning
it out both because *I* want to use this in my next project and I hope
you may find it useful.

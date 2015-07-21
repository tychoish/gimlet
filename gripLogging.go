package gimlet

import (
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/tychoish/grip"
)

type appLogging struct {
	*grip.Journaler
}

func newAppLogger() *appLogging {
	l := &appLogging{grip.NewJournaler("gimlet+negroni")}

	// default to whatever grip's standard logger does.
	l.PreferFallback = grip.PrefersFallback()

	return l
}

func (self *appLogging) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	self.Infof("Started %s %s", r.Method, r.URL.Path)

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	self.Infof("Completed %v %s in %v", res.Status(), http.StatusText(res.Status()), time.Since(start))
}

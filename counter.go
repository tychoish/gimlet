package gimlet

import (
	"context"
	"net/http"
)

var jobIDSource <-chan int

func init() {
	jobIDSource = func() <-chan int {
		out := make(chan int, 50)
		go func() {
			var jobID int
			for {
				jobID++
				out <- jobID
			}
		}()
		return out
	}()
}

// getNumber is a source of safe monotonically increasing integers
// for use in request ids.
func getNumber() int {
	return <-jobIDSource
}

func setRequestID(r *http.Request, id int) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), requestIDKey, id))
}

// GetRequestID returns the unique (monotonically increaseing) ID of
func GetRequestID(r *http.Request) int {
	if rv := r.Context().Value(requestIDKey); rv != nil {
		if id, ok := rv.(int); ok {
			return id
		}
	}

	return 0
}

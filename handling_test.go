package gimlet

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tychoish/grip"
	"github.com/tychoish/grip/send"
)

type writeResponseBaseFunc func(http.ResponseWriter, int, interface{})
type writeResponseFunc func(http.ResponseWriter, interface{})

func TestResponseWritingFunctions(t *testing.T) {
	assert := assert.New(t)
	baseCases := map[OutputFormat]writeResponseBaseFunc{
		JSON:   WriteJSONResponse,
		BINARY: WriteBinaryResponse,
		YAML:   WriteYAMLResponse,
		TEXT:   WriteTextResponse,
		HTML:   WriteHTMLResponse,
	}

	for of, wf := range baseCases {
		for _, code := range []int{200, 400, 500} {
			r := httptest.NewRecorder()
			wf(r, code, "")
			assert.Equal(code, r.Code)

			header := r.Header()
			assert.Len(header, 1)
			ct, ok := header["Content-Type"]
			assert.True(ok)
			assert.Equal(ct, []string{of.ContentType()})

			body := r.Body.Bytes()
			assert.True(len(body) < 4)
		}
	}
}

func TestSerializationErrors(t *testing.T) {
	assert := assert.New(t)

	baseCases := map[OutputFormat]writeResponseBaseFunc{
		JSON: WriteJSONResponse,
		YAML: WriteYAMLResponse,
		CSV:  WriteCSVResponse,
	}

	for _, wf := range baseCases {
		r := httptest.NewRecorder()

		wf(r, http.StatusOK, struct{ Foo chan struct{} }{Foo: make(chan struct{})})
		assert.Equal(r.Code, http.StatusInternalServerError)

		wf(r, http.StatusOK, errors.New("foo"))
		assert.Equal(r.Code, http.StatusInternalServerError)
	}
}

func TestResponsesWritingHelpers(t *testing.T) {
	assert := assert.New(t)
	testTable := map[int]map[OutputFormat]writeResponseFunc{
		http.StatusOK: {
			JSON:   WriteJSON,
			BINARY: WriteBinary,
			YAML:   WriteYAML,
			TEXT:   WriteText,
			HTML:   WriteHTML,
		},
		http.StatusBadRequest: {
			JSON:   WriteJSONError,
			BINARY: WriteBinaryError,
			YAML:   WriteYAMLError,
			TEXT:   WriteTextError,
			HTML:   WriteHTMLError,
		},
		http.StatusInternalServerError: {
			JSON:   WriteJSONInternalError,
			BINARY: WriteBinaryInternalError,
			YAML:   WriteYAMLInternalError,
			TEXT:   WriteTextInternalError,
			HTML:   WriteHTMLInternalError,
		},
	}

	for status, cases := range testTable {
		for of, wf := range cases {
			r := httptest.NewRecorder()
			wf(r, []struct{}{})
			assert.Equal(status, r.Code)
			ct, ok := r.Header()["Content-Type"]
			assert.True(ok)
			assert.Equal(ct, []string{of.ContentType()})
			body := r.Body.Bytes()
			assert.True(len(body) < 4)
		}
	}
}

func TestBytesConverter(t *testing.T) {
	assert := assert.New(t)

	cases := [][]interface{}{
		{fmt.Sprintf("%v", http.DefaultClient), http.DefaultClient},
		{"gimlet", "gimlet"},
		{"gimlet", errors.New("gimlet")},
		{"gimlet", []byte("gimlet")},
		{"GET", get},
		{"gimlet", bytes.NewBufferString("gimlet")},
	}

	for _, c := range cases {
		if !assert.Len(c, 2) {
			continue
		}

		out := c[0].(string)
		in := c[1]

		buf := &bytes.Buffer{}
		_, err := writePayload(buf, in)
		assert.NoError(err)
		assert.Equal([]byte(out), buf.Bytes())
	}

	buf := &bytes.Buffer{}
	_, err := writePayload(buf, []string{"gimlet", "gimlet"})
	assert.NoError(err)
	assert.Equal([]byte("gimlet\ngimlet"), buf.Bytes())
}

type mangledResponseWriter struct{ *httptest.ResponseRecorder }

func (r mangledResponseWriter) Write(b []byte) (int, error) { return 0, errors.New("always errors") }

func TestWriteResponseErrorLogs(t *testing.T) {
	defer func(s send.Sender) {
		grip.SetGlobalLogger(grip.NewLogger(s))
	}(grip.Sender())
	assert := assert.New(t)
	sender := send.MakeInternalLogger()
	rw := mangledResponseWriter{httptest.NewRecorder()}
	grip.SetGlobalLogger(grip.NewLogger(sender))

	assert.False(sender.HasMessage())
	writeResponse(JSON, rw, 200, []byte("foo"))
	assert.True(sender.HasMessage())
}

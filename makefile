projectPath := github.com/tychoish/gimlet
coverageFile := coverage.out

deps:
	go get github.com/codegangsta/negroni
	go get github.com/gorilla/mux
	go get github.com/phyber/negroni-gzip/gzip
	go get github.com/tychoish/grip
	go get github.com/tylerb/graceful

test-deps:deps
	go get github.com/stretchr/testify/suite
	go get github.com/alecthomas/gometalinter
	gometalinter --install

build:deps
	go build -v

lint:
	gometalinter --disable=gotype --deadline=20s

test:
	go test -v -covermode=count -coverprofile=${coverageFile} ${projectPath}
	go tool cover -func=${coverageFile}

coverage-report:test
	go tool cover -html=${coverageFile}

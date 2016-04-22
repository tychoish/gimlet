PACKAGES := ./

deps:
	go get github.com/codegangsta/negroni
	go get github.com/gorilla/mux
	go get github.com/phyber/negroni-gzip/gzip
	go get github.com/tychoish/grip
	go get github.com/tylerb/graceful

test-deps:deps
	go get -u gopkg.in/check.v1

build:deps
	go build -v

lint:
	gofmt -l $(PACKAGES)
	go vet $(PACKAGES)

test:build lint
	go test -cover -v -check.v $(PACKAGES)


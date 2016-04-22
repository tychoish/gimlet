deps:
	go get github.com/codegangsta/negroni
	go get github.com/gorilla/mux
	go get github.com/phyber/negroni-gzip/gzip
	go get github.com/tychoish/grip
	go get github.com/tylerb/graceful

test-deps:deps
	go get gopkg.in/check.v1
	-go get github.com/alecthomas/gometalinter
	-gometalinter --install --update

build:deps
	go build -v

lint:
	gofmt -l ./
	go vet ./
	-gometalinter --disable=gotype --deadline=20s

test:build lint
	go test -cover -v -check.v

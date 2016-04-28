# project configuration
name := gimlet
buildDir := build
packages :=
projectPath := github.com/tychoish/$(name)


# declaration of dependencies
lintDeps := github.com/alecthomas/gometalinter
testDeps := github.com/stretchr/testify
deps += github.com/codegangsta/negroni
deps += github.com/gorilla/mux
deps += github.com/phyber/negroni-gzip/gzip
deps += github.com/tychoish/grip
deps += github.com/tylerb/graceful

# implementation details for being able to lazily
gopath := $(shell go env GOPATH)
deps := $(addprefix $(gopath)/src/,${deps})
lintDeps := $(addprefix $(gopath)/src/,${lintDeps})
testDeps := $(addprefix $(gopath)/src/,${testDeps})
$(gopath)/src/%:
	go get $(subst $(gopath)/src/,,$@)
# end dependency installation tools


# userfacing targets for basic build/test/lint operations
.PHONY:build test lint coverage-report
build:deps
	@mkdir -p $@
	go build
test:test-deps
	go test -v ./...
lint:
	$(gopath)/bin/gometalinter --deadline=20s --disable=gotype ./...
coverage:$(foreach target,$(packages),$(buildDir)/coverage.$(target).out) $(buildDir)/coverage.out $(buildDir)
coverage-report:$(foreach target,$(packages),coverage-report-$(target))
# end front-ends

# implementation for package coverage
coverage-%:$(buildDir)/coverage.%.out
$(buildDir)/coverage.out:test-deps #(buildDir)
# this special target is only needed because there are tests in the root package.
	go test -v -covermode=count -coverprofile=$@ $(projectPath)
	[ -f $@ ] && go tool cover -func=$@ | sed 's%${projectPath}/%%' | column -t
$(buildDir)/coverage.%.out:% test-deps
	go test -v -covermode=count -coverprofile=$@ $(projectPath)/$<
	[ -f $@ ] && go tool cover -func=$@ | sed 's%${projectPath}/%%' | column -t
coverage-report-%:$(buildDir)/coverage.%.out
	[ -f $< ] && go tool cover -html=$<
# end coverage rports


# targets to install dependencies
deps:$(deps)
test-deps:$(testDeps)
lint-deps:$(lintDeps)
	gometalinter --install
clean:
	rm -rf $(deps) $(lintDeps) $(testDeps)

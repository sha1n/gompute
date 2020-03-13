GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOFILES := $(shell find . -type f -name '*.go' -not -path './vendor/*')


default: go-get go-format go-lint go-test

go-get:
	@echo "  >  Fetching deps..."
	@GOPATH=$(GOPATH) go mod tidy

go-test:
	@echo "  >  Running tests..."
	go test -v `go list ./...`


go-lint:
	@echo "  >  Linting source files..."
	gofmt -d $(GOFILES)

go-format:
	@echo "  >  Formating source files..."
	gofmt -s -w $(GOFILES)

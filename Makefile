NOCOL=\x1b[0m
GREEN=\x1b[32;01m
RED=\x1b[31;01m
YELLOW=\x1b[33;01m

define print_title
	@echo "---"
	@echo "--- $(GREEN)$1$(NOCOL)"
	@echo "---"
endef


default: get lint format test

get:
	go get -t ./...

test:
	$(call print_title, Running tests...)
	go test -v `go list ./...`


prepare:
	$(call print_title,Preparing go dependencies...)
	dep ensure -v


format:
	$(call print_title,Formatting go sources...)
	gofmt -s -w .


lint:
	$(call print_title,Lint...)
	gofmt -d .

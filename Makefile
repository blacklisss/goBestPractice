HAS_LINT := $(shell command -v golangci-lint;)
HAS_IMPORTS := $(shell command -v goimports;)

bootstrap:
ifndef HAS_LINT
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.32.2
endif
ifndef HAS_IMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	 go test -race ./...

.PHONY: imports
imports: bootstrap
	@echo "+ $@"
	@go list -f '"goimports -w {{.Dir}}"' ${GO_PKG} | xargs -L 1 sh -c

.PHONY: lint
lint: bootstrap
	@echo "+ $@"
	golangci-lint run ./...

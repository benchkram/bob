SHELL := /usr/bin/env bash

.PHONY: help
help:
	@egrep -h '\s#\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: build # build and run the tool
	@./run $(filter-out $@,$(MAKECMDGOALS))

.PHONY: build
build: # build the tool
	@go build -o ./run

.PHONY: install
install: # install the tool
	@go install

.PHONY: test
test: # run tests
	@go test -v ./...

.PHONY: lint
lint: # lint code
	@CGO_ENABLED=0 golangci-lint run --timeout=10m0s

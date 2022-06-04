VERSION := "v0.0.0"
SHELL := /usr/bin/env bash

.PHONY: help
help:
	@egrep -h '\s#\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: # install the tool
	@go install -tags dev -ldflags="-X 'main.Version=${VERSION}'"

.PHONY: install
install-prod: # install the tool
	@go install -ldflags="-X 'main.Version=${VERSION}'"


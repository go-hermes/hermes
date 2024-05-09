.PHONY: test
.ONESHELL:
SHELL = /bin/bash
.SHELLFLAGS = -ec

MAKEFLAGS += -s

test:
	@echo "Running tests..."
	go test -coverprofile=cp.out ./...

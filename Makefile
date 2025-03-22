.PHONY: help deps

all: help

deps:
	go get -u ./...
	go mod tidy
	go mod vendor

help:
	@echo "Available commands:"
	@echo "  make deps - Update dependencies and create vendor directory"

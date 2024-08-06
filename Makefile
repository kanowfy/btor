MAIN_PKG_PATH := .
BINARY_NAME := btor

## help: display help messages
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## no-dirty: checks that there are no uncommitted change in the tracked files
.PHONY: no-dirty
no-dirty:
	git diff --exit-code

## tidy: format code and tidy mod file
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy
	
## audit: run various quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go test -race -vet=off ./...

## test: run all tests
.PHONY: test
test:
	go test -race ./...

## test/cover: run test and display cover profile
.PHONY: test/cover
test/cover:
	go test -v -race -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out


## build: build the program
.PHONY: build
build:
	go build -o bin/${BINARY_NAME} ${MAIN_PKG_PATH}

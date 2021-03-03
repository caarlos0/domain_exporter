SOURCE_FILES?=./...
TEST_PATTERN?=.
TEST_OPTIONS?=

export GO111MODULE := on

setup:
	go mod tidy
.PHONY: setup

test:
	go test $(TEST_OPTIONS) -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

cover: test
	go tool cover -html=coverage.out
.PHONY: cover

fmt:
	gofumpt -w .
.PHONY: fmt

.DEFAULT_GOAL := test

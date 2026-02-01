APP=epagent
BIN_DIR=bin

.PHONY: setup dev test lint typecheck build check release

setup:
	go mod tidy

dev:
	go run ./cmd/epagent collect --once

test:
	go test ./...

lint:
	@unformatted=$$(gofmt -l .); if [ -n "$$unformatted" ]; then echo "gofmt needed:"; echo "$$unformatted"; exit 1; fi
	go vet ./...

typecheck:
	go test ./... -run TestDoesNotExist

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/epagent

check: lint typecheck test build

release:
	@echo "Release is handled via git tags and GitHub Releases."

PREFIX?=$(shell pwd)

NAME := storms

PKG := gitlab.com/crusoeenergy/island/storage/storms/storms-cli/cmd/$(NAME) # VHENG -- TODO. this should later point to the non-cli

BUILDDIR := ${PREFIX}/dist
# Set any default go build tags
BUILDTAGS :=

GOLANGCI_VERSION = v2.2.1
TOOLS_VERSION = v0.7.0
GO_ACC_VERSION = latest
GOTESTSUM_VERSION = latest
GOCOVER_VERSION = latest

.PHONY: ci
ci: test-ci build-deps lint-ci ## Runs test, build-deps, lint

.PHONY: build-deps
build-deps: ## Install build dependencies
	@echo "==> $@"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_VERSION}

.PHONY: get-aliaslint
get-aliaslint:
	@echo "==> $@"
	@go mod download gitlab.com/crusoeenergy/tools@${TOOLS_VERSION}
	@cd `go env GOMODCACHE`/gitlab.com/crusoeenergy/tools@${TOOLS_VERSION}; go build -o ${PREFIX}/aliaslint.so -buildmode=plugin aliaslint/plugin/aliaslint.go

.PHONY: test-ci
test-ci: ## Runs the go tests with additional options for a CI environment
	@echo "==> $@"
	@go mod tidy
	@git diff --exit-code go.mod go.sum # fail if go.mod is not tidy
	@go install github.com/ory/go-acc@${GO_ACC_VERSION}
	@go install gotest.tools/gotestsum@${GOTESTSUM_VERSION}
	@go install github.com/boumenot/gocover-cobertura@${GOCOVER_VERSION}
	@gotestsum --junitfile tests.xml --raw-command -- go-acc -o coverage.out ./... -- -json -tags "$(BUILDTAGS)" -race -v
	@go tool cover -func=coverage.out
	@gocover-cobertura < coverage.out > coverage.xml

.PHONY: lint-ci
lint-ci: get-aliaslint ## Verifies `golangci-lint` passes and outputs in CI-friendly format
	@echo "==> $@"
	@golangci-lint version
	@

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

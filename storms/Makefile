PREFIX?=$(shell pwd)

# TODO(template) update this to match your executable's name
NAME := storage-management-service
# TODO(template) update this to point to your project
PKG := gitlab.com/crusoeenergy/island/storage/storms/storage-management-service/cmd/$(NAME) # VHENG -- TODO. this should later point to the non-cli

BUILDDIR := ${PREFIX}/dist
# Set any default go build tags
BUILDTAGS :=

.PHONY: build
build: ## Builds the executable and places it in the build dir
	@go build -o ${BUILDDIR}/${NAME} ${PKG}

.PHONY: run
	@go run cmd/storms/main.go --config dev/config.yaml

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

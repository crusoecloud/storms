PREFIX?=$(shell pwd)

NAME := storms-cli
PKG := gitlab.com/crusoeenergy/island/storage/storms/storms-cli/cmd/$(NAME)

BUILDDIR := ${PREFIX}/dist

.PHONY: build
build: prep ## Builds the executable and places it in the build dir
	@go build -o ${BUILDDIR}/${NAME} ${PKG}

.PHONY: install
install: prep ## Builds and installs the executable on PATH
	@go install ${PKG}
	
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

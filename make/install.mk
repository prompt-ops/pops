GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)

VERSION ?= dev

.PHONY: install
install: build
	@echo "Installing Prompt-Ops version $(VERSION)..."
	@cp dist/pops-$(GOOS)-$(GOARCH) /usr/local/bin/pops
	@echo "Installation complete."
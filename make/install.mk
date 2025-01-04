GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)

.PHONY: install
install: build
	@echo "Installing Prompt-Ops..."
	@cp dist/pops-$(GOOS)-$(GOARCH) /usr/local/bin/pops
	@echo "Installation complete."
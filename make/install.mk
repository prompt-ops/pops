GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)

VERSION ?= dev

.PHONY: install
install: build
	@echo "Installing Prompt-Ops version $(VERSION)..."
	
	@echo "Deleting existing installation..."
	@rm -f /usr/local/bin/pops
	
	@echo "Copying new installation..."
	@cp dist/pops-$(GOOS)-$(GOARCH) /usr/local/bin/pops
	
	@echo "Installation complete."
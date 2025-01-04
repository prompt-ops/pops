GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)

GO111MODULE ?= on
CGO_ENABLED ?= 0

.PHONY: build
build:
	@echo "Building Prompt-Ops..."
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "GOPATH: $(GOPATH)"
	@echo "GO111MODULE: $(GO111MODULE)"
	@echo "CGO_ENABLED: $(CGO_ENABLED)"
	@go build -ldflags="-s -w" -o dist/pops-$(GOOS)-$(GOARCH)
	@echo "Build complete."
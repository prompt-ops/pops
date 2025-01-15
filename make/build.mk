GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)

GO111MODULE ?= on
CGO_ENABLED ?= 0
VERSION ?= dev

.PHONY: build
build:
	@echo "Building Prompt-Ops..."
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "GOPATH: $(GOPATH)"
	@echo "GO111MODULE: $(GO111MODULE)"
	@echo "CGO_ENABLED: $(CGO_ENABLED)"
	@echo "VERSION: $(VERSION)"
	@go build -ldflags="-s -w -X github.com/prompt-ops/pops/cmd/pops/app.version=$(VERSION)" -o dist/pops-$(GOOS)-$(GOARCH) cmd/pops/main.go
	@echo "Build complete."
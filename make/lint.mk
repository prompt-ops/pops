GOOS ?= $(shell go env GOOS)

ifeq ($(GOOS),windows)
   GOLANGCI_LINT:=golangci-lint.exe
else
   GOLANGCI_LINT:=golangci-lint
endif

.PHONY: lint
lint:
	@echo "Running Go linter..."
	@$(GOLANGCI_LINT) run --fix --timeout 5m
	@echo "Linting complete."
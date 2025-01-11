.PHONY: unit-test

unit-test:
	@echo "Running unit tests..."
	@go test ./... -v
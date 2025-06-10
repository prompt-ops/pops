.PHONY: arch-test arch-demo arch-validate

arch-test:
	@echo "Running architecture tests..."
	@echo "Testing domain layer..."
	@go test ./pkg/domain -v -cover
	@echo "Testing service layer..."
	@go test ./pkg/services -v -cover
	@echo "Testing adapter layer..."
	@go test ./pkg/adapters -v -cover
	@echo "Architecture tests complete."

arch-demo:
	@echo "Running architecture demonstration..."
	@go run cmd/demo/main.go
	@echo "Architecture demo complete."

arch-validate:
	@echo "Running architecture validation..."
	@go run cmd/test-harness/main.go > /dev/null
	@echo "✅ Architecture validation passed."

# Development build with full validation
dev-build: arch-validate arch-test lint build
	@echo "🎉 Development build complete with full validation."

# Safe development install (parallel to production)
dev-install: dev-build
	@echo "Installing development build..."
	@cp dist/pops-$(GOOS)-$(GOARCH) /usr/local/bin/pops-dev
	@echo "✅ Installed as 'pops-dev' for safe testing."
	@echo "   Test with: pops-dev version"

# Full validation before release
pre-release: clean arch-validate arch-test unit-test lint
	@echo "✅ Pre-release validation complete."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist/
	@echo "Clean complete."
.PHONY: generate-cli-docs
generate-cli-docs:
	@echo "Generating CLI docs for Prompt-Ops..."
	@go run cmd/docgen/main.go docs
	@echo "Generation complete."
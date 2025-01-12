OUTPUT_PATH ?= docs

.PHONY: generate-cli-docs
generate-cli-docs:
	@echo "Generating CLI docs for Prompt-Ops..."
	@go run cmd/docgen/main.go $(OUTPUT_PATH)
	@echo "Generation complete. Docs generated at $(OUTPUT_PATH)"
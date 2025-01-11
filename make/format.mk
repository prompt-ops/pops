.PHONY: organize-imports

organize-imports:
	@echo "Organizing imports and formatting code with goimports..."
	@goimports -w .
	@echo "Formatting complete."
# Makefile for go-dnp3
# Documentation and project management tasks only
#
# Note: This project is in the research and architecture phase.
# No implementation code exists yet.

.PHONY: help
.DEFAULT_GOAL := help

# Colors
BOLD := $(shell tput bold)
GREEN := $(shell tput setaf 2)
YELLOW := $(shell tput setaf 3)
NC := $(shell tput sgr0) # No Color

## help - Display this help message
help:
	@echo ""
	@echo "${BOLD}${GREEN}go-dnp3${NC} - Project Management Commands"
	@echo ""
	@echo "${BOLD}Documentation:${NC}"
	@grep -E '^[##[:space:]]*[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  ${BOLD}%-22s${NC} %s\n", $$1, $$2}'
	@echo ""

## docs - Build documentation site (when implemented)
docs:
	@echo "${YELLOW}Note: Documentation build not yet implemented${NC}"
	@echo "Documentation is available in the docs/ directory"

## lint-docs - Lint documentation files
lint-docs:
	@echo "Checking documentation..."
	@find docs -name "*.md" -exec echo "Checking: {}" \; -exec grep -l "TODO\|FIXME\|XXX" {} \; || true
	@echo "Documentation lint complete"

## spell-check - Check documentation for spelling errors
spell-check:
	@echo "${YELLOW}Note: Spell checking not yet configured${NC}"

## links - Check documentation links
links:
	@echo "${YELLOW}Note: Link checking not yet configured${NC}"

## format-docs - Format documentation files
format-docs:
	@echo "${YELLOW}Note: Documentation formatting not yet configured${NC}"

## test - Run tests (placeholder)
test:
	@echo "${YELLOW}No tests available - project is in research phase${NC}"

## vet - Run go vet (placeholder)
vet:
	@echo "${YELLOW}No code to vet - project is in research phase${NC}"

## build - Build the project (placeholder)
build:
	@echo "${YELLOW}No code to build - project is in research phase${NC}"

## clean - Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf docs/_build site/
	@rm -f coverage.* *.out
	@echo "Clean complete"

## bootstrap - Initialize project structure
bootstrap:
	@echo "Verifying project structure..."
	@for dir in docs/architecture docs/adr docs/research docs/specifications docs/roadmap examples cmd internal pkg test scripts .github; do \
		if [ -d "$$dir" ]; then \
			echo "  [OK] $$dir"; \
		else \
			echo "  [MISSING] $$dir"; \
		fi \
	done
	@echo ""
	@echo "Bootstrap verification complete"

## list-files - List all project files
list-files:
	@echo "Project structure:"
	@find . -type f -not -path '*/.git/*' | sort

## commit-msg - Validate commit message format
commit-msg:
	@echo "Commit message validation (conventional commits):"
	@echo "  feat: add new feature"
	@echo "  fix: fix existing issue"
	@echo "  docs: documentation changes"
	@echo "  refactor: code refactoring"
	@echo "  test: test updates"
	@echo "  chore: maintenance tasks"

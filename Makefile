# Makefile for go-dnp3
# Build, test, and project management tasks
#
# This project contains partial implementation with 41 Go source files.
# Integration between layers is incomplete.

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

## test - Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage report:"
	go tool cover -func=coverage.out

## vet - Run go vet for static analysis
vet:
	@echo "Running go vet..."
	go vet ./...

## build - Build all packages
build:
	@echo "Building..."
	go build ./...

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

## kde-start - Start KDE runtime with quick preflight and bootstrap verification
kde-start:
	@echo "=== KDE Runtime Startup ==="
	@echo ""
	@echo "Step 1: Bootstrap Status Check"
	@python3 .kde/bootstrap/status.py || true
	@echo ""
	@echo "Step 2: Environment Preflight (Quick)"
	@python3 .kde/bootstrap/gates.py --quick
	@echo ""
	@echo "Step 3: Starting KDE Engine..."
	@python3 -c "import sys; sys.path.insert(0, '.kde'); from runtime.runtime import demo; demo()"

## kde-check - Run full preflight check with bootstrap status
kde-check:
	@echo "=== KDE Full Check ==="
	@echo ""
	@echo "Step 1: Bootstrap Status"
	@python3 .kde/bootstrap/status.py
	@echo ""
	@echo "Step 2: Environment Gates (Full)"
	@python3 .kde/bootstrap/gates.py --full

## kde-quick - Run quick preflight check (skip slow checks)
kde-quick:
	@echo "=== KDE Quick Check ==="
	@python3 .kde/bootstrap/gates.py --quick

## kde-status - Show bootstrap status (no environment checks)
kde-status:
	@python3 .kde/bootstrap/status.py

## kde-watch - Watch bootstrap status continuously
kde-watch:
	@python3 .kde/bootstrap/status.py --watch --interval 5

## kde-help - Show KDE-specific help
kde-help:
	@echo ""
	@echo "${BOLD}${GREEN}KDE Runtime Commands${NC}"
	@echo ""
	@echo "${BOLD}make kde-start${NC}   - Start KDE runtime (bootstrap + preflight + engine)"
	@echo "${BOLD}make kde-check${NC}    - Full bootstrap and environment check"
	@echo "${BOLD}make kde-quick${NC}    - Quick preflight check (skip slow checks)"
	@echo "${BOLD}make kde-status${NC}   - Show bootstrap status (no environment checks)"
	@echo "${BOLD}make kde-watch${NC}    - Watch bootstrap status continuously"
	@echo ""
	@echo "Direct Commands:"
	@echo "  python3 .kde/bootstrap/gates.py --quick  # Quick preflight"
	@echo "  python3 .kde/bootstrap/gates.py --full   # Full preflight"
	@echo "  python3 .kde/bootstrap/status.py         # Bootstrap status"
	@echo ""
	@echo "For more details on KDE Runtime, see:"
	@echo "  .kde/README.md"
	@echo "  laboratory/investigations/"
	@echo ""

## commit-msg - Validate commit message format
commit-msg:
	@echo "Commit message validation (conventional commits):"
	@echo "  feat: add new feature"
	@echo "  fix: fix existing issue"
	@echo "  docs: documentation changes"
	@echo "  refactor: code refactoring"
	@echo "  test: test updates"
	@echo "  chore: maintenance tasks"

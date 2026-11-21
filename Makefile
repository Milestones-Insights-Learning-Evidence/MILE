# MILE (Milestones, Insight, Learning & Evidence) - Build, Test, Boundary & License Verification Makefile
# Usage:
#   make build           - Build Community Edition binary (./bin/mile)
#   make build-ee        - Build Enterprise Edition binary with -tags ee (./bin/mile-ee)
#   make test            - Run unit tests (Community build)
#   make test-ee         - Run unit tests (Enterprise -tags ee build)
#   make test-race       - Run unit tests with Go race detector (Community & EE)
#   make vet             - Run go vet on both Community and Enterprise build tags
#   make staticcheck     - Run staticcheck on both Community and Enterprise build tags
#   make fmt-check       - Verify all Go files are formatted with gofmt
#   make boundary-check  - Verify default Community build imports zero packages from ee/
#   make license-check   - Verify all go.mod dependencies match approved permissive allowlist
#   make generate        - Run SQLc code generation (Phase 2+)
#   make ui              - Build static SvelteKit frontend bundle (Phase 4+)
#   make clean           - Remove build artifacts in ./bin

GO          ?= go
GOFMT       ?= gofmt
SQLC        ?= sqlc
GO_LICENSES ?= go-licenses
STATICCHECK ?= staticcheck
BIN_DIR     := bin
BINARY      := $(BIN_DIR)/mile
BINARY_EE   := $(BIN_DIR)/mile-ee
MODULE      := github.com/Milestones-Insights-Learning-Evidence/MILE

# Approved permissive runtime module allowlist (see docs/governance/DEPENDENCY_LICENSES.md)
APPROVED_MODULES := \
	github.com/Milestones-Insights-Learning-Evidence/MILE \
	github.com/prometheus/client_golang \
	github.com/beorn7/perks \
	github.com/cespare/xxhash/v2 \
	github.com/munnerz/goautoneg \
	github.com/prometheus/client_model \
	github.com/prometheus/common \
	github.com/prometheus/procfs \
	golang.org/x/sys \
	google.golang.org/protobuf

.PHONY: all build build-ee test test-ee test-race vet staticcheck fmt-check boundary-check license-check generate ui clean

all: fmt-check vet staticcheck boundary-check license-check test test-ee build build-ee

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -o $(BINARY) ./cmd/mile

build-ee:
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -tags ee -o $(BINARY_EE) ./cmd/mile

test:
	$(GO) test ./...

test-ee:
	$(GO) test -tags ee ./...

test-race:
	$(GO) test -race ./...
	$(GO) test -race -tags ee ./...

vet:
	$(GO) vet ./...
	$(GO) vet -tags ee ./...

# Files behind //go:build ee are invisible to a default-tag analysis, so the ee build
# is checked in a second pass. CI mirrors this with two staticcheck-action steps.
staticcheck:
	@command -v $(STATICCHECK) >/dev/null 2>&1 || \
		{ echo "ERROR: staticcheck not installed. Install it with: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
	$(STATICCHECK) ./...
	$(STATICCHECK) -tags ee ./...
	@echo "OK: staticcheck passed for Community and Enterprise build tags."

fmt-check:
	@unformatted=$$($(GOFMT) -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "ERROR: The following Go files are not formatted with gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "OK: All Go files are properly formatted."

boundary-check:
	@ee_deps=$$($(GO) list -deps ./cmd/mile | grep '^$(MODULE)/ee' || true); \
	if [ -n "$$ee_deps" ]; then \
		echo "ERROR: Community build imports proprietary ee/ packages:"; \
		echo "$$ee_deps"; \
		exit 1; \
	fi
	@echo "OK: Community build dependency graph contains zero ee/ packages."

# Gates the modules actually linked into the shipped binaries, for both editions and
# for the GOOS/GOARCH that is distributed (linux/amd64 containers). `go list -m all`
# must NOT be used here: module-graph pruning still reports test-only dependencies of
# dependencies (testify, kingpin, json-iterator, ...) that are never compiled into
# bin/mile, so it fails the gate on software MILE does not distribute. That is exactly
# the distributed-vs-build-tool distinction in docs/governance/DEPENDENCY_LICENSES.md section 2.
LINKED_MODULES = $(sort $(shell GOOS=linux GOARCH=amd64 $(GO) list -deps -f "{{with .Module}}{{.Path}}{{end}}" ./cmd/mile) $(shell GOOS=linux GOARCH=amd64 $(GO) list -deps -tags ee -f "{{with .Module}}{{.Path}}{{end}}" ./cmd/mile))

license-check:
	@for mod in $(LINKED_MODULES); do \
		allowed=0; \
		for ok in $(APPROVED_MODULES); do \
			if [ "$$mod" = "$$ok" ]; then allowed=1; break; fi; \
		done; \
		if [ "$$allowed" -eq 0 ]; then \
			echo "ERROR: Module '$$mod' is linked into a shipped binary but is not on the approved permissive allowlist."; \
			echo "Verify its SPDX license (NFR-03), then update docs/governance/DEPENDENCY_LICENSES.md section 4 and APPROVED_MODULES in this Makefile."; \
			exit 1; \
		fi; \
	done
	@if command -v $(GO_LICENSES) >/dev/null 2>&1; then \
		$(GO_LICENSES) check ./cmd/mile --disallowed_types=forbidden,restricted,reciprocal,unknown; \
	else \
		echo "INFO: go-licenses not installed locally; CI runs it. Install: go install github.com/google/go-licenses@latest"; \
	fi
	@echo "OK: All distributed Go modules satisfy the permissive license allowlist (NFR-03)."

generate:
	@if command -v $(SQLC) >/dev/null 2>&1; then \
		$(SQLC) generate; \
	else \
		echo "INFO: sqlc not installed; skipping code generation (configured in Phase 2)."; \
	fi

ui:
	@if [ -d "ui/node_modules" ]; then \
		cd ui && npm run build; \
	else \
		echo "INFO: ui/node_modules not present; UI build activates in Phase 4."; \
	fi

clean:
	rm -rf $(BIN_DIR)

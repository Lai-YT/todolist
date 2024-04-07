GO ?= go
GOFMT ?= gofumpt
GO_VERSION=$(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
PACKAGES ?= $(shell $(GO) list ./...)
GOFILES := $(shell find . -name "*.go")
TESTTAGS ?= "-test.shuffle=on"
COVERPROFILE ?= coverage.out
COVEREXCLUDE ?= "$$^"

.PHONY: test
test:
	$(GO) test $(TESTTAGS) -v $(PACKAGES)

.PHONY: test-coverage
test-coverage:
	$(GO) test $(TESTTAGS) -v $(PACKAGES) -coverprofile=/tmp/$(COVERPROFILE)
	cat /tmp/$(COVERPROFILE) | grep -v -E $(COVEREXCLUDE) > $(COVERPROFILE)
	$(GO) tool cover -func=$(COVERPROFILE)

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: vet
vet:
	$(GO) vet $(PACKAGES)

.PHONY: lint
lint:
	@hash staticcheck > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install honnef.co/go/tools/cmd/staticcheck; \
	fi
	staticcheck $(PACKAGES)

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -error $(GOFILES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -w $(GOFILES)

.PHONY: tools
tools:
	$(GO) install mvdan.cc/gofumpt@latest; \
	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest; \
	$(GO) install github.com/client9/misspell/cmd/misspell@latest;

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  test            Run tests"
	@echo "  test-coverage   Run tests with coverage"
	@echo "  fmt             Format code"
	@echo "  fmt-check       Check code format"
	@echo "  vet             Run go vet"
	@echo "  lint            Run staticcheck"
	@echo "  misspell-check  Check spelling"
	@echo "  misspell        Fix spelling"
	@echo "  tools           Install tools"
	@echo "  help            Show this help message"

GO ?= go
GOFMT ?= gofmt "-s"
GO_VERSION=$(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
PACKAGES ?= $(shell $(GO) list ./...)
VETPACKAGES ?= $(shell $(GO) list ./... | grep -v /examples/)
GOFILES := $(shell find . -name "*.go")
TESTFOLDER := $(shell $(GO) list ./...)
TESTTAGS ?=

.PHONY: test
test:
	echo "mode: count" > coverage.out
	for d in $(TESTFOLDER); do \
		$(GO) test $(TESTTAGS) -v -covermode=count -coverprofile=profile.out $$d > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "setup failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done

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
	$(GO) vet $(VETPACKAGES)

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/client9/misspell/cmd/misspell@v0.3.4; \
	fi
	misspell -error $(GOFILES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/client9/misspell/cmd/misspell@v0.3.4; \
	fi
	misspell -w $(GOFILES)

.PHONY: tools
tools:
	$(GO) install github.com/client9/misspell/cmd/misspell@v0.3.4
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

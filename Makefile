GO ?= go
GOCACHE ?= /tmp/go-cache
GOPATH ?= /tmp/go
GOMODCACHE ?= $(GOPATH)/pkg/mod

# Run the full test suite with sandbox-friendly cache locations.
test:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) $(GO) test ./...

# Run only the strategies package tests (fast, no Fiber dependency).
test-strategies:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) $(GO) test ./strategies

.PHONY: test test-strategies

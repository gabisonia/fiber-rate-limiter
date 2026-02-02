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

# Run full suite while forcing recompilation (does not disable cache, but ignores it).
test-nocache:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) $(GO) test -count=1 -a ./...

.PHONY: test test-strategies test-nocache

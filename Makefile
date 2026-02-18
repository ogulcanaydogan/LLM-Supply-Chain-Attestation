GO ?= go

.PHONY: init tidy build test demo attest sign verify gate report

init:
	$(GO) run ./cmd/llmsa init

tidy:
	$(GO) mod tidy

build:
	$(GO) build ./cmd/llmsa

test:
	$(GO) test ./...

demo:
	$(MAKE) -C examples/tiny-rag demo

attest:
	$(MAKE) -C examples/tiny-rag attest

sign:
	$(MAKE) -C examples/tiny-rag sign

verify:
	$(MAKE) -C examples/tiny-rag verify

gate:
	$(MAKE) -C examples/tiny-rag gate

report:
	$(MAKE) -C examples/tiny-rag report

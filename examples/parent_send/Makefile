.PHONY: help run run-debug test clean

REPO_ROOT := $(abspath $(CURDIR)/../..)
GOCACHE ?= $(REPO_ROOT)/.gocache

help:
	@printf "%s\n" \
		"Targets:" \
		"  make run        - run the example" \
		"  make run-debug  - run with -tags=debug (enables axy contract asserts)" \
		"  make test       - run go test ./... from repo root" \
		"  make clean      - remove local caches" \
		"" \
		"Variables:" \
		"  GOCACHE=$(GOCACHE)"

run:
	cd "$(REPO_ROOT)" && GOCACHE="$(GOCACHE)" go run ./examples/parent_send

run-debug:
	cd "$(REPO_ROOT)" && GOCACHE="$(GOCACHE)" go run -tags=debug ./examples/parent_send

test:
	cd "$(REPO_ROOT)" && GOCACHE="$(GOCACHE)" go test ./...

clean:
	@rm -rf "$(REPO_ROOT)/.gocache" "$(REPO_ROOT)/examples/parent_send/.gocache"


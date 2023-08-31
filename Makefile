VERSION=1.0.36
#VER_MAJ_MIN      := $(shell git describe --tags)
VER_MAJ_MIN      := $(subst ., ,$(VERSION))
VERSION        := $(word 1,$(VER_MAJ_MIN))
MAJOR          := $(word 2,$(VER_MAJ_MIN))
MINOR          := $(word 3,$(VER_MAJ_MIN))

.PHONY: all

all: set-tag push-tag

set-tag:
	git tag "v$(VERSION).$(MAJOR).$(MINOR)"
push-tag:
	git push --tags

go-lint:
	golangci-lint run

vendor:
	go mod vendor

tidy:
	go mod tidy

VERSION=1.0.14
#VER_MAJ_MIN      := $(shell git describe --tags)
VER_MAJ_MIN      := $(subst ., ,$(shell git describe --tags))
VERSION        := $(word 1,$(VER_MAJ_MIN))
MAJOR          := $(word 2,$(VER_MAJ_MIN))
MINOR          := $(word 3,$(VER_MAJ_MIN))
NEW_MINOR := $(shell expr "$(MINOR)" + 1)

.PHONY: all

all: set-tag push-tag

set-tag:
	git tag "$(VERSION).$(MAJOR).$(NEW_MINOR)"
push-tag:
	git push --tags
.PHONY: all

all: set-tag push-tag

set-tag:
	git tag "v1.0.13"
push-tag:
	git push --tags
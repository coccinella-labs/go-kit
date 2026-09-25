.PHONY: all build test lint clean install run docs

VERSION ?= 0.1.0
GOCMD = go
GOFLAGS = -ldflags "-X github.com/coccinella-labs/go-kit/cmd/kit.Version=$(VERSION)"
BINARY = kit

all: build

build:
	$(GOCMD) build $(GOFLAGS) -o bin/$(BINARY) ./cmd/kit

test:
	$(GOCMD) test -v ./...

lint:
	$(GOCMD) vet ./...

clean:
	rm -rf bin/

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/

run:
	$(GOCMD) run ./cmd/kit

docs:
	$(GOCMD) run ./tools/docsgen

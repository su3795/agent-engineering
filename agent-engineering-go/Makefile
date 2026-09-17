# agent-engineering-go Makefile

BINARY := agent-engineering
PKG := ./cmd/agent-engineering

.PHONY: build run test verify tenant fmt clean

build:
	go build -o $(BINARY) $(PKG)

run:
	go run $(PKG) run --engine mock

verify:
	go run $(PKG) verify

tenant:
	go run $(PKG) tenant --strict

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -f $(BINARY)

.PHONY: build build-current build-all relay mcp tidy

GOHOSTOS := $(shell go env GOOS)
GOHOSTARCH := $(shell go env GOARCH)
SERVER_BIN := bin/$(GOHOSTOS)-$(GOHOSTARCH)/figma-mcp-design
RELAY_BIN := bin/$(GOHOSTOS)-$(GOHOSTARCH)/figma-mcp-relay
ifeq ($(GOHOSTOS),windows)
	SERVER_BIN := bin/$(GOHOSTOS)-$(GOHOSTARCH)/figma-mcp-design.exe
	RELAY_BIN := bin/$(GOHOSTOS)-$(GOHOSTARCH)/figma-mcp-relay.exe
endif

build: build-current build-all

build-current:
	go build ./...
	CGO_ENABLED=0 go build -o $(SERVER_BIN) ./cmd/mcp-server
	CGO_ENABLED=0 go build -o $(RELAY_BIN) ./cmd/relay

build-all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/figma-mcp-design ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/figma-mcp-relay ./cmd/relay
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/figma-mcp-design ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/figma-mcp-relay ./cmd/relay
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/figma-mcp-design.exe ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/figma-mcp-relay.exe ./cmd/relay
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o bin/windows-arm64/figma-mcp-design.exe ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o bin/windows-arm64/figma-mcp-relay.exe ./cmd/relay
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/figma-mcp-design ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/figma-mcp-relay ./cmd/relay
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/figma-mcp-design ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/figma-mcp-relay ./cmd/relay

relay:
	./$(RELAY_BIN)

mcp:
	./$(SERVER_BIN)

tidy:
	go mod tidy

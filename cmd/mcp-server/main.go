package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"figma-mcp-design/internal/figma"
	"figma-mcp-design/internal/mcp"
	"figma-mcp-design/internal/relay"
)

func main() {
	serverHost := flag.String("server", "localhost", "WebSocket relay host. Use localhost for ws://localhost:<port>; any other value uses wss://<server>.")
	port := flag.Int("port", 3055, "WebSocket relay port when --server=localhost")
	wsURL := flag.String("ws-url", "", "Full WebSocket relay URL. Overrides --server and --port.")
	embeddedRelay := flag.Bool("embedded-relay", true, "Start a local WebSocket relay in this process when using localhost and the port is available.")
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	url := *wsURL
	if url == "" {
		if *serverHost == "localhost" {
			url = fmt.Sprintf("ws://localhost:%d", *port)
		} else {
			url = fmt.Sprintf("wss://%s", *serverHost)
		}
	}

	if *embeddedRelay && shouldStartEmbeddedRelay(*serverHost, *wsURL) {
		startEmbeddedRelay(*port)
	}

	client := figma.NewClient(url)
	client.ConnectAsync()

	server := mcp.NewServer(client)
	if err := server.Serve(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("mcp server stopped: %v", err)
	}
}

func shouldStartEmbeddedRelay(serverHost string, wsURL string) bool {
	if wsURL != "" {
		return strings.HasPrefix(wsURL, "ws://localhost:") || strings.HasPrefix(wsURL, "ws://127.0.0.1:")
	}
	return serverHost == "localhost" || serverHost == "127.0.0.1"
}

func startEmbeddedRelay(port int) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("embedded relay not started on %s: %v; using an existing relay if available", addr, err)
		return
	}
	_ = listener.Close()

	go func() {
		log.Printf("starting embedded WebSocket relay on %s", addr)
		if err := relay.ListenAndServe(addr); err != nil {
			log.Printf("embedded relay stopped: %v", err)
		}
	}()
}

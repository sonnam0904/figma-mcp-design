package main

import (
	"log"
	"os"

	"figma-mcp-design/internal/relay"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3055"
	}

	if err := relay.ListenAndServe(":" + port); err != nil {
		log.Fatalf("relay stopped: %v", err)
	}
}

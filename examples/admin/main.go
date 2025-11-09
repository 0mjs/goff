package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		dbPath   = flag.String("db", "flags.db", "Path to SQLite database")
		yamlPath = flag.String("yaml", "flags.yaml", "Path to YAML output file")
		port     = flag.String("port", "8081", "Admin server port")
	)
	flag.Parse()

	// Ensure YAML directory exists
	if dir := filepath.Dir(*yamlPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create YAML directory: %v", err)
		}
	}

	store, err := NewFlagStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	server := NewAdminServer(store, *yamlPath, *port)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

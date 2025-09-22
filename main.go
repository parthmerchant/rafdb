package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"rafdb/internal/server"
	"rafdb/internal/storage"
)

func main() {
	// Initialize the database
	db := storage.NewDatabase()

	// Load existing data from disk if available
	if err := db.LoadFromDisk(); err != nil {
		log.Printf("Warning: Could not load existing data: %v", err)
	}

	// Start the HTTP server
	srv := server.NewServer(db)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")

		// Save data to disk before shutdown
		if err := db.SaveToDisk(); err != nil {
			log.Printf("Error saving data to disk: %v", err)
		}

		srv.Shutdown()
		os.Exit(0)
	}()

	log.Println("Starting RAFDB server on :8080")
	srv.Start(":8080")
}

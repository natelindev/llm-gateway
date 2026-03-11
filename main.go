package main

import (
	"log"
	"os"

	"llm-gateway/internal/registry"
	"llm-gateway/internal/router"
)

func main() {
	// Read listen port from env, default to 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize in-memory model registry.
	reg := registry.NewRegistry()

	// Configure router and middleware.
	r := router.Setup(reg)

	log.Printf("LLM Gateway starting on port: %s", port)
	log.Printf("Health endpoint: http://localhost:%s/health", port)
	log.Printf("Prometheus metrics: http://localhost:%s/metrics", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start service: %v", err)
	}
}

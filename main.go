package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"quickship_api/pkg/aggregator"

	"github.com/gorilla/mux"
)

// --- Configuration ---
const defaultPort = ":8080"
const dashboardPort = ":3000" // React will typically run on 3000

// GetCartSummary is the HTTP entry point.
func GetCartSummary(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	productID := vars["sku"]

	log.Printf("Request received for SKU: %s", productID)

	// Execute the core concurrency logic
	summary, err := aggregator.FanOutAndAggregate(productID)

	if err != nil {
		http.Error(w, "Failed to aggregate data due to internal system error", http.StatusInternalServerError)
		return
	}

	// Finalize response time metrics
	elapsed := time.Since(start)
	summary.TotalTimeMs = elapsed.Milliseconds()

	log.Printf("Processed SKU: %s in %dms. Errors: %d", productID, summary.TotalTimeMs, len(summary.ServiceErrors))

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Enable CORS using the CORS_ORIGIN environment variable.
func enableCORS(next http.Handler) http.Handler {
    // Read the allowed origin from the environment variable.
    // If not set (like in local dev), it defaults to localhost:3000.
    allowedOrigin := os.Getenv("CORS_ORIGIN")
    if allowedOrigin == "" {
        allowedOrigin = "http://localhost:3000" // Default for local dev
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set the Access-Control-Allow-Origin header using the value from the environment.
        w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// main sets up the router and starts the HTTP server.
func main() {
	serverPort := os.Getenv("PORT") 
    if serverPort == "" {
        serverPort = defaultPort
    } else {
        serverPort = ":" + serverPort
    }
	
	r := mux.NewRouter()

	// Use CORS middleware
	r.Use(enableCORS)

	// Define the route with a variable SKU path.
	r.HandleFunc("/cart/summary/{sku}", GetCartSummary).Methods("GET")
	
	// // Add a simple route for testing connectivity
	// r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	// 	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	// }).Methods("GET")


	log.Printf("Server starting on http://localhost%s", serverPort)

	// Start the server (blocking)
	if err := http.ListenAndServe(serverPort, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
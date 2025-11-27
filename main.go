package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"quickship_api/pkg/aggregator"

	"github.com/gorilla/mux"
)

// --- Configuration ---
const serverPort = ":8080"
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

// Enable CORS for development between Go and React
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost"+dashboardPort)
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
	r := mux.NewRouter()

	// Use CORS middleware
	r.Use(enableCORS)

	// Define the route with a variable SKU path.
	r.HandleFunc("/cart/summary/{sku}", GetCartSummary).Methods("GET")
	
	// Add a simple route for testing connectivity
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}).Methods("GET")


	log.Printf("Server starting on http://localhost%s", serverPort)

	// Start the server (blocking)
	if err := http.ListenAndServe(serverPort, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
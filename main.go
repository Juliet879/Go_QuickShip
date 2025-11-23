package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

/*
################################################################################
QUICKSHIP: GO CONCURRENT MICROSERVICE AGGREGATION SYSTEM

This application implements an HTTP endpoint (/cart/summary/{sku}) that fetches 
multiple independent pieces of data (Price, Stock, Promotion) from simulated 
external microservices concurrently using Go's Goroutines and Channels.

The Fan-Out/Fan-In pattern is used to reduce the overall response latency from
the sequential sum of all service latencies (650ms) to the duration of the 
slowest service (400ms).

### API Endpoint: /cart/summary/{sku}

**Input Example (cURL Request):**
```bash
curl http://localhost:8080/cart/summary/PRODUCT-XYZ-123
```

**Output Example (JSON Response):**
```json
{
  "product_id": "PRODUCT-XYZ-123",
  "final_price": 49.99,
  "available_stock": 120,
  "promotion_message": "Buy 1 Get 1 Half Off!",
  "total_time_ms": 405  // Expected value close to 400ms (the slowest service)
}
```
################################################################################
*/

// --- Configuration ---
const serverPort = ":8080"

// --- Data Structures ---

// CartSummary represents the final aggregated data returned to the client.
type CartSummary struct {
	ProductID   string  `json:"product_id"`
	Price       float64 `json:"final_price"`
	Stock       int     `json:"available_stock"`
	Promotion   string  `json:"promotion_message"`
	TotalTimeMs int64   `json:"total_time_ms"` // For performance benchmarking
}

// FetchResult is the common structure for results passed through the channel.
type FetchResult struct {
	Key   string      // Identifier for the data type (e.g., "price", "stock")
	Value interface{} // The actual data fetched (type asserted during Fan-In)
	Error error       // Any error encountered during the fetch
}

// ServiceFn is a standardized type for all external service mock functions.
// This standardization is key for abstracting the execution pattern.
type ServiceFn func(productID string) (interface{}, error)

// --- Mock External Service Implementations (The Fan-Out Workers) ---

// fetchPriceSimulates fetches pricing data. Second slowest (200ms).
func fetchPriceSimulates(productID string) (interface{}, error) {
	const delay = 200 * time.Millisecond
	time.Sleep(delay)
	// In a real scenario, business logic and error checks would happen here.
	return 49.99, nil 
}

// fetchInventorySimulates fetches stock data. The bottleneck (400ms).
func fetchInventorySimulates(productID string) (interface{}, error) {
	const delay = 400 * time.Millisecond
	time.Sleep(delay)
	return 120, nil 
}

// fetchPromotionsSimulates fetches promotion data. The fastest (50ms).
func fetchPromotionsSimulates(productID string) (interface{}, error) {
	const delay = 50 * time.Millisecond
	time.Sleep(delay)
	return "Buy 1 Get 1 Half Off!", nil
}

// --- Concurrency Abstraction ---

// executeService is a consolidated helper function that runs a ServiceFn 
// in a goroutine, manages the WaitGroup, and sends the result to the channel.
// This function implements the standard "Worker" pattern in the Fan-Out architecture.
func executeService(
	key string, 
	productID string, 
	fn ServiceFn, 
	resultChan chan<- FetchResult, 
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	value, err := fn(productID)

	// Send the result or error back to the main goroutine (the aggregator).
	resultChan <- FetchResult{
		Key:   key,
		Value: value,
		Error: err,
	}
}

// fanOutAndAggregate executes the core Fan-Out/Fan-In pattern. 
// It launches all workers concurrently and then reads from the channel until completion.
// 
func fanOutAndAggregate(productID string) (CartSummary, error) {
	// Define the map of services to execute. This is the central configuration for Fan-Out.
	services := map[string]ServiceFn{
		"price":     fetchPriceSimulates,
		"stock":     fetchInventorySimulates,
		"promotion": fetchPromotionsSimulates,
	}
	
	// Prepare synchronization tools
	var wg sync.WaitGroup
	numWorkers := len(services)
	// Buffered channel size prevents workers from blocking.
	resultChan := make(chan FetchResult, numWorkers) 

	// 1. Fan-Out: Launch all workers concurrently.
	wg.Add(numWorkers)
	for key, fn := range services {
		// Use a local copy of key for the goroutine closure
		key := key 
		go executeService(key, productID, fn, resultChan, &wg)
	}

	// 2. Fan-In Coordination: Wait for all workers and then close the channel.
	// This separate goroutine prevents a deadlock in the main function.
	go func() {
		wg.Wait()
		close(resultChan) // Signals the aggregation loop (step 3) to terminate.
	}()

	// 3. Aggregation (The Fan-In): Read results until the channel closes.
	summary := CartSummary{ProductID: productID}
	
	for res := range resultChan {
		if res.Error != nil {
			log.Printf("Error fetching data for key '%s': %v. Skipping.", res.Key, res.Error)
			// A fail-soft approach: log the error and continue aggregating other data.
			continue
		}

		// Use type assertion to assign the correct data to the summary struct.
		switch res.Key {
		case "price":
			summary.Price = res.Value.(float64)
		case "stock":
			summary.Stock = res.Value.(int)
		case "promotion":
			summary.Promotion = res.Value.(string)
		default:
			log.Printf("Received unexpected key: %s", res.Key)
		}
	}
	
	// At this point, all workers have completed and all results have been processed.
	return summary, nil
}

// --- HTTP Handler ---

// GetCartSummary is the HTTP entry point. It handles request parsing, 
// timing, and final response formatting.
func GetCartSummary(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	productID := vars["sku"]

	log.Printf("Request received for SKU: %s", productID)
	
	// Execute the core concurrency logic
	summary, err := fanOutAndAggregate(productID)
	
	if err != nil {
		// In this specific implementation, fanOutAndAggregate returns an error only 
		// if a severe logical failure occurred within the concurrency setup itself.
		http.Error(w, "Failed to aggregate data due to internal error", http.StatusInternalServerError)
		return
	}

	// Finalize response time metrics
	elapsed := time.Since(start)
	summary.TotalTimeMs = elapsed.Milliseconds()

	log.Printf("Successfully processed SKU: %s in %dms.", productID, summary.TotalTimeMs)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// main sets up the router and starts the HTTP server.
func main() {
	r := mux.NewRouter()
	
	// Define the route with a variable SKU path.
	r.HandleFunc("/cart/summary/{sku}", GetCartSummary).Methods("GET")

	log.Printf("Server starting on http://localhost%s", serverPort)
	
	// Start the server (blocking)
	if err := http.ListenAndServe(serverPort, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
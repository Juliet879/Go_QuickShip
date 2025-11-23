package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// TestGetCartSummary performs an integration test of the refactored handler
// and verifies that the total time taken is approximately the time of the
// slowest concurrent worker (400ms), proving the Fan-Out/Fan-In pattern worked.
func TestGetCartSummary(t *testing.T) {
	testSKU := "TEST-SKU-REFAC"

	// 1. Setup: Create a request and a ResponseRecorder
	req, err := http.NewRequest("GET", fmt.Sprintf("/cart/summary/%s", testSKU), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Setup Gorilla Mux to correctly handle path variables
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/cart/summary/{sku}", GetCartSummary).Methods("GET")
	
	// Execute the handler through the router
	router.ServeHTTP(rr, req)

	// 2. Verification of Status Code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
		return
	}

	// 3. Verification of Content and Performance
	var summary CartSummary
	if err := json.NewDecoder(rr.Body).Decode(&summary); err != nil {
		t.Fatalf("Could not decode response body: %v", err)
	}

	// Define expected values
	expectedPrice := 49.99
	expectedStock := 120
	expectedPromo := "Buy 1 Get 1 Half Off!"
	
	// Concurrency Check Range: Slowest service is 400ms. We verify the time is near 400ms.
	minExpectedTimeMs := int64(390) 
	maxExpectedTimeMs := int64(450) 

	// Check aggregated data correctness
	if summary.ProductID != testSKU {
		t.Errorf("Data check failed: Expected ProductID %s, got %s", testSKU, summary.ProductID)
	}
	if summary.Price != expectedPrice {
		t.Errorf("Data check failed: Expected Price %.2f, got %.2f", expectedPrice, summary.Price)
	}
	if summary.Stock != expectedStock {
		t.Errorf("Data check failed: Expected Stock %d, got %d", expectedStock, summary.Stock)
	}
	if summary.Promotion != expectedPromo {
		t.Errorf("Data check failed: Expected Promotion %s, got %s", expectedPromo, summary.Promotion)
	}
	
	// Check the core benefit: Concurrency/Performance
	if summary.TotalTimeMs < minExpectedTimeMs || summary.TotalTimeMs > maxExpectedTimeMs {
		t.Errorf("Performance test failed: Expected total time between %dms and %dms, got %dms. Sequential time is 650ms.",
			minExpectedTimeMs, maxExpectedTimeMs, summary.TotalTimeMs)
		t.Logf("This result verifies the Fan-Out/Fan-In pattern by confirming execution time is dominated by the slowest worker.")
	}
}
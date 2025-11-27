package services

import (
	"context"
	"fmt"
	"time"

	"quickship_api/pkg/models"
)

// simulatedCall simulates a network call with a delay and a potential error.
func simulatedCall(ctx context.Context, delay time.Duration, productID string, fail bool, value interface{}) (interface{}, error) {
	select {
	case <-time.After(delay):
		if fail {
			return nil, fmt.Errorf("backend server responded with 503")
		}
		return value, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// fetchPriceSimulates fetches pricing data. Second slowest (200ms).
func FetchPriceSimulates(ctx context.Context, productID string) (interface{}, error) {
    delay := 200 * time.Millisecond // Default delay

    // Trigger Timeout for specific SKU ---
    if productID == "TEST-SKU-TIMEOUT" {
        // Set the delay longer than the 450ms global timeout in aggregator.go
        delay = 500 * time.Millisecond 
    }
    // --------------------------------------------------

    return simulatedCall(ctx, delay, productID, false, 49.99)
}

// fetchInventorySimulates fetches stock data. The bottleneck (400ms).
func FetchInventorySimulates(ctx context.Context, productID string) (interface{}, error) {
	return simulatedCall(ctx, 400*time.Millisecond, productID, false, 120)
}

// fetchPromotionsSimulates fetches promotion data. The fastest (50ms).
func FetchPromotionsSimulates(ctx context.Context, productID string) (interface{}, error) {
	return simulatedCall(ctx, 50*time.Millisecond, productID, false, "Buy 1 Get 1 Half Off!")
}

// ServiceMap holds the configuration for Fan-Out services.
var ServiceMap = map[string]models.ServiceFn{
	"price":     FetchPriceSimulates,
	"stock":     FetchInventorySimulates,
	"promotion": FetchPromotionsSimulates,
}
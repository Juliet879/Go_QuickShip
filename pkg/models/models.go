package models

import (
	"context"
	"fmt"
)

// --- Custom Error Types ---

// ServiceError provides better context for failures.
type ServiceError struct {
	ServiceKey string `json:"service_key"`
	Message    string `json:"message"`
	Timeout    bool   `json:"timeout"` // True if the error was due to an explicit service-level timeout
	Err        error  `json:"-"`       // Exclude internal error details from JSON response
}

func (e *ServiceError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("Service '%s' timed out: %s (%v)", e.ServiceKey, e.Message, e.Err)
	}
	return fmt.Sprintf("Service '%s' failed: %s (%v)", e.ServiceKey, e.Message, e.Err)
}

// --- Status Types ---

// Status represents the status of an individual service fetch.
type Status string

const (
	StatusSuccess  Status = "SUCCESS"
	StatusTimeout  Status = "TIMEOUT"
	StatusError    Status = "ERROR"
	StatusFallback Status = "FALLBACK"
)

// --- Data Structures ---

// CartSummary represents the final aggregated data returned to the client.
type CartSummary struct {
	ProductID         string         `json:"product_id"`
	Price             float64        `json:"final_price"`
	PriceStatus       Status         `json:"price_status"`
	Stock             int            `json:"available_stock"`
	StockStatus       Status         `json:"stock_status"`
	Promotion         string         `json:"promotion_message"`
	PromotionStatus   Status         `json:"promotion_status"`
	TotalTimeMs       int64          `json:"total_time_ms"`
	ServiceErrors     []*ServiceError `json:"service_errors,omitempty"`
}

// FetchResult is the common structure for results passed through the channel.
type FetchResult struct {
	Key   string      // Identifier for the data type (e.g., "price", "stock")
	Value interface{} // The actual data fetched
	Status Status
	Error *ServiceError
}

// ServiceFn is a standardized type for all external service mock functions.
type ServiceFn func(ctx context.Context, productID string) (interface{}, error)
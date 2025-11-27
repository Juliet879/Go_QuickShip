package aggregator

import (
	"context"
	"log"
	"sync"
	"time"

	"quickship_api/pkg/models"
	"quickship_api/pkg/services"
)

const aggregationTimeout = 450 * time.Millisecond // Global timeout for all services

// executeService is a consolidated helper function that runs a ServiceFn 
// in a goroutine, manages the WaitGroup, and sends the result to the channel.
func executeService(
	ctx context.Context,
	key string,
	productID string,
	fn models.ServiceFn,
	resultChan chan<- models.FetchResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	value, err := fn(ctx, productID)
	result := models.FetchResult{Key: key, Value: value}

	if err != nil {
		serviceErr := &models.ServiceError{
			ServiceKey: key,
			Err:        err,
			Message:    "External service call failed",
		}

		if err == context.DeadlineExceeded {
			serviceErr.Timeout = true
			serviceErr.Message = "Service execution exceeded the global aggregation timeout"
			result.Status = models.StatusTimeout
		} else {
			result.Status = models.StatusError
		}

		result.Error = serviceErr
		resultChan <- result
		return
	}

	result.Status = models.StatusSuccess
	resultChan <- result
}

// FanOutAndAggregate executes the core Fan-Out/Fan-In pattern with a timeout context.
func FanOutAndAggregate(productID string) (models.CartSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), aggregationTimeout)
	defer cancel()

	numWorkers := len(services.ServiceMap)
	resultChan := make(chan models.FetchResult, numWorkers)
	var wg sync.WaitGroup

	// 1. Fan-Out: Launch all workers concurrently, passing the context.
	wg.Add(numWorkers)
	for serviceName, serviceFunc := range services.ServiceMap {
		currentName := serviceName
		currentFunc := serviceFunc
		go executeService(ctx, currentName, productID, currentFunc, resultChan, &wg)
	}

	// 2. Fan-In Coordination: Wait for all workers and then close the channel.
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 3. Aggregation (The Fan-In): Read results and apply fallbacks.
	summary := models.CartSummary{ProductID: productID}
	var serviceErrors []*models.ServiceError

	// Define fallback values
	const (
		fallbackPrice     = 0.00
		fallbackStock     = 0
		fallbackPromotion = "No promotions available"
	)

	for res := range resultChan {
		if res.Error != nil {
			log.Printf("Failure for service '%s': %v", res.Key, res.Error)
			serviceErrors = append(serviceErrors, res.Error)
		}

		// Apply fallback or assign success value
		switch res.Key {
		case "price":
			if res.Status == models.StatusSuccess {
				summary.Price = res.Value.(float64)
				summary.PriceStatus = models.StatusSuccess
			} else {
				summary.Price = fallbackPrice
				summary.PriceStatus = models.StatusFallback
			}
		case "stock":
			if res.Status == models.StatusSuccess {
				summary.Stock = res.Value.(int)
				summary.StockStatus = models.StatusSuccess
			} else {
				summary.Stock = fallbackStock
				summary.StockStatus = models.StatusFallback
			}
		case "promotion":
			if res.Status == models.StatusSuccess {
				summary.Promotion = res.Value.(string)
				summary.PromotionStatus = models.StatusSuccess
			} else {
				summary.Promotion = fallbackPromotion
				summary.PromotionStatus = models.StatusFallback
			}
		default:
			log.Printf("Received unexpected key: %s", res.Key)
		}
	}

	summary.ServiceErrors = serviceErrors
	return summary, nil
}
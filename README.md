QuickShip: High-Speed E-commerce Data Aggregation
QuickShip is a high-speed e-commerce data aggregation service that demonstrates how to concurrently fetch product data (price, inventory, promotions) from multiple microservices to deliver lightning-fast API responses.
It uses Go’s Fan-Out / Fan-In pattern to slash end-to-end latency.

🚀 Why QuickShip Exists

Modern e-commerce systems rely on many services for real-time product info.
Calling them sequentially is too slow.
QuickShip solves this by running all service calls in parallel, returning results as fast as the slowest service.

🎯 Performance Breakdown
🐢 Sequential Execution (Slow)
50ms + 200ms + 400ms = 650ms

⚡ Concurrent Execution (QuickShip Speed)
~400ms (dictated by slowest service)

Service Latencies
Service	Latency
fetchPromotionsSimulates	50ms
fetchPriceSimulates	200ms
fetchInventorySimulates	400ms
🧩 Architecture Diagram (Fan-Out / Fan-In)
             ┌────────────────┐
Request  →   │ GetCartSummary │
             └───────┬────────┘
                     │
             (Fan-Out: Launch workers)
                     ▼
      ┌──────────────┬──────────────┐
      ▼              ▼              ▼
 Promotions     Price Service    Inventory
   Worker          Worker          Worker
 (50ms)           (200ms)         (400ms)
      └──────────────┬──────────────┘
                     │
             (Fan-In: Combine)
                     ▼
          Final Cart Summary JSON

🧑‍💻 Refactored Code Structure
Component	Purpose
GetCartSummary	HTTP endpoint; coordinates request/response.
fanOutAndAggregate	Core engine for concurrency + aggregation.
executeService	Standard wrapper for running service functions safely.
ServiceFn	Type definition for easily pluggable services.
fetch*Simulates	Mock versions simulating real service delays.
🛠️ Prerequisites

Go 1.18+

Gorilla Mux

go get github.com/gorilla/mux

▶️ Running the Server

Place main.go and main_test.go in your project folder.

Start the app:

http://localhost:8080


Open in browser or Postman:

http://localhost:8080

⚡ Testing the Speed

Run:

```bash
curl http://localhost:8080/cart/summary/SKU-REFAC-TEST
```

Expected Response

total_time_ms should be ~400ms:

{
  "product_id": "SKU-REFAC-TEST",
  "final_price": 49.99,
  "available_stock": 120,
  "promotion_message": "Buy 1 Get 1 Half Off!",
  "total_time_ms": 405
}
```

🧪 Running Unit Tests
go test -v .


Tests verify:

endpoints return correct data

concurrency reduces total execution time

📂 Project Structure
QuickShip/
├── main.go
├── main_test.go
├── go.mod
└── go.sum

📜 License
You are free to use, modify, and distribute this project.

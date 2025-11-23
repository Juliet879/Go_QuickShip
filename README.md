QuickShip: High-Speed E-commerce Data Aggregation

This project demonstrates a critical technique for high-performance e-commerce APIs: concurrently aggregating product data (like price, inventory, and promotions) from multiple independent microservices to ensure a blazing fast checkout experience. It utilizes the Go Fan-Out/Fan-In pattern to achieve maximum speed.

🎯 Performance Metrics

The code executes all simulated services in parallel, slashing response latency:

Service

Latency

fetchPromotionsSimulates

50ms

fetchPriceSimulates

200ms

fetchInventorySimulates

400ms (Bottleneck)

Sequential Time (Slow): Sum of all service latencies (200ms + 400ms + 50ms = 650ms).

Concurrent Time (QuickShip Speed): Approximately the time of the slowest service (~400ms).

🧑‍💻 Refactored Code Structure

The application's logic is structured for clarity and easy maintenance:

Function/Type

Responsibility

GetCartSummary

The public HTTP endpoint; handles request parsing, timing, and response.

fanOutAndAggregate

The Core Engine: Manages the concurrent execution (Fan-Out) and gathers all results (Fan-In).

executeService

Standardized wrapper for running any microservice worker and safely reporting results.

ServiceFn

Type definition for external service functions, ensuring easy integration of new services.

fetch*Simulates

Mock implementations simulating external I/O delays and data return.

🚀 Setup and Running

Prerequisites

Go (v1.18 or later)

The Gorilla Mux router:

go get [github.com/gorilla/mux](https://github.com/gorilla/mux)


Running the Server

Place the code in main.go and main_test.go into a new Go project directory.

Run the application:

go run main.go


The server will start on http://localhost:8080.

Testing the Speed

Query the endpoint to observe the concurrent speedup:

curl http://localhost:8080/cart/summary/SKU-REFAC-TEST


Expected Output: The total_time_ms should be close to 400ms.

{
  "product_id": "SKU-REFAC-TEST",
  "final_price": 49.99,
  "available_stock": 120,
  "promotion_message": "Buy 1 Get 1 Half Off!",
  "total_time_ms": 405  
}


Running the Unit Test

The included test file, main_test.go, automatically verifies both data correctness and the critical performance gain.

go test -v .

## QuickShip: High-Speed E-commerce Data Aggregation

QuickShip is a high-speed e-commerce data aggregation service that concurrently fetches product data (price, inventory, promotions) from multiple microservices to deliver lightning-fast API responses.  
It uses Go’s **Fan-Out / Fan-In** concurrency pattern to drastically cut latency.

(./assets/dashboard.png)

![Toolkit guide ](https://docs.google.com/document/d/1dPYXK5CbDxIvK8qaDKIiKjNl-UrFTXKzh3q8zu7GJJ8/edit?tab=t.0)
---

## 🚀 Why QuickShip Exists

Modern e-commerce platforms depend on several microservices to compute real-time product information.  
Calling them **one-by-one is too slow**.

QuickShip solves this by:

- Running all service calls **in parallel**
- Returning results as fast as **the slowest service**
- Ensuring consistent, low-latency API responses

---

## 🎯 Performance Breakdown

### 🐢 Sequential Execution (Slow)
```

50ms + 200ms + 400ms = 650ms

```

### ⚡ Concurrent Execution (QuickShip Speed)
```

~400ms (determined by the slowest service)

```

### Service Latencies

| Service                   | Latency |
|---------------------------|---------|
| fetchPromotionsSimulates | 50ms    |
| fetchPriceSimulates      | 200ms   |
| fetchInventorySimulates  | 400ms   |

---

## 🧩 Architecture Diagram (Fan-Out / Fan-In)

         ┌────────────────┐
```

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

````

---

## 🛠️ Prerequisites

- Go **1.18+**
- Gorilla Mux router  
  ```bash
  go get github.com/gorilla/mux
````

---
````
## ▶️ Running the Server

Place `main.go` and `main_test.go` in your project directory.

Start the application:

```bash
go run main.go
```

Open in browser or Postman:

```
http://localhost:8080
```

---

## ⚡ Testing the Speed

Run:

```bash
curl http://localhost:8080/cart/summary/SKU-REFAC-TEST
```

### Expected Response (~400ms total):

```json
{
  "product_id": "SKU-REFAC-TEST",
  "final_price": 49.99,
  "available_stock": 120,
  "promotion_message": "Buy 1 Get 1 Half Off!",
  "total_time_ms": 405
}
```

---

## 🧪 Running Unit Tests

```bash
go test -v .
```

Tests verify:

* Correct data returned
* Concurrency reduces execution time

---

## 📂 Project Structure

```
QuickShip/
├── main.go
├── main_test.go
├── go.mod
└── go.sum
```

---

## 📜 License

📜 License
You are free to use, modify, and distribute this project.

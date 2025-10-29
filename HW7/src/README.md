# E-commerce API with Payment Bottleneck Simulation

This is a RESTful API for managing a product catalog and processing orders. It demonstrates how a payment processing bottleneck impacts system performance during high-load scenarios like flash sales.

## Features

### Product Management
*   `GET /products`: Retrieve a list of products with filtering support
*   `GET /products/{id}`: Retrieve a specific product by ID
*   `POST /products`: Create a new product (admin functionality)

### Order Processing
*   `POST /orders`: Create and process a new order with payment verification
*   Payment processing simulates a **3-second external payment provider delay**
*   Uses semaphore pattern to limit concurrent payment processing to **5 simultaneous transactions**

### System Monitoring
*   `GET /stats`: Real-time system performance metrics including:
    *   Total orders processed
    *   Success/failure rates
    *   Timeout counts
    *   Current queue depth
    *   Maximum queue depth observed

### Data Generation
*   Generates 100,000 products in memory on startup
*   Product names: "Product [AppName] [ID]"
*   Brands: `["Apple", "Samsung", "Google", "Microsoft", "Sony", "Logitech"]`
*   Categories: `["Electronics", "Books", "Home", "Clothing", "Groceries", "Toys"]`
*   Random descriptions and prices

### Search Logic
*   **Fixed Computation Time:** Searches check maximum 100 products
*   **Filtering:** Case-insensitive partial match for brands, exact match for categories
*   **Result Limiting:** Maximum 20 products per response

## The Bottleneck

### Normal Operations (5 orders/sec capacity)
- Payment processor handles 5 concurrent transactions
- Each payment verification takes 3 seconds
- System handles ~5 orders/second smoothly

### Flash Sale Scenario (60 orders/sec load)
- 60 orders/second arrive
- Only 5 can process simultaneously (semaphore limit)
- Queue builds up rapidly
- Orders timeout after 30 seconds of waiting
- **Result: System degradation and customer timeouts**

### Implementation Details
```go
// Buffered channel as semaphore - limits to 5 concurrent payments
paymentSemaphore = make(chan int, 5)

// Acquire semaphore (blocks if 5 payments already processing)
paymentSemaphore <- 1

// Simulate 3-second payment verification
time.Sleep(3 * time.Second)

// Release semaphore
<-paymentSemaphore
```

## Getting Started

### Building and Running

```bash
# Build Docker image
docker build -t ecommerce-api .

# Run container
docker run -p 8080:8080 ecommerce-api
```

### Generate from OpenAPI (Optional)
```bash
oapi-codegen \
  --package api \
  -o src/api/server.gen.go \
  --generate types,server \
  ecommerce.yaml
```

## API Endpoints

Base URL: `http://localhost:8080`

### `GET /products`

Retrieves a list of products with optional filtering.

**Query Parameters:**
*   `brands` (string, optional): Filter by brand (case-insensitive, partial match)
*   `categories` (string, optional): Filter by category (case-insensitive, exact match)

**Example Request:**
```bash
curl "http://localhost:8080/products?brands=Sam&categories=Home"
```

**Example Response:**
```json
{
  "products": [
    {
      "id": "5",
      "name": "Product Go 5",
      "category": "Home",
      "brand": "Samsung",
      "description": "Innovative home solution",
      "price": 789.12
    }
  ],
  "total_found": 12,
  "search_time": "1.2ms"
}
```

### `GET /products/{id}`

Retrieves a specific product by ID.

**Example Request:**
```bash
curl http://localhost:8080/products/1
```

**Example Response:**
```json
{
  "id": "1",
  "name": "Product Apple 1",
  "category": "Electronics",
  "brand": "Apple",
  "description": "Premium technology device",
  "price": 123.45
}
```

### `POST /orders`

Creates and processes a new order with payment verification.

**Request Body:**
```json
{
  "product_id": "1",
  "quantity": 2
}
```

**Example Request:**
```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "product_id": "1",
  "quantity": 2
}' http://localhost:8080/orders
```

**Success Response (201 Created):**
```json
{
  "order_id": "1",
  "product_id": "1",
  "quantity": 2,
  "total_price": 246.90,
  "status": "completed",
  "processing_time": "3.005s",
  "queue_time": "2ms"
}
```

**Timeout Response (503 Service Unavailable):**
```json
{
  "order_id": "150",
  "product_id": "1",
  "quantity": 1,
  "total_price": 123.45,
  "status": "timeout - payment processor overloaded",
  "processing_time": "30.001s",
  "queue_time": "25s"
}
```

### `GET /stats`

Returns real-time system performance metrics.

**Example Request:**
```bash
curl http://localhost:8080/stats
```

**Example Response:**
```json
{
  "total_orders": 1500,
  "successful_orders": 450,
  "failed_orders": 1050,
  "timeout_orders": 1045,
  "current_queue": 55,
  "max_queue_seen": 1200,
  "success_rate": "30.00%"
}
```

## Load Testing

Test the bottleneck with a load testing tool like `locust` or `wrk`:

### Using curl in a loop (simple test)
```bash
# Normal load - 5 requests/sec
for i in {1..100}; do
  curl -X POST -H "Content-Type: application/json" \
    -d '{"product_id": "1", "quantity": 1}' \
    http://localhost:8080/orders &
  sleep 0.2
done
```

### Using Apache Bench
```bash
# Flash sale simulation - 60 requests/sec
ab -n 3600 -c 60 -p order.json -T application/json \
  http://localhost:8080/orders
```

Where `order.json` contains:
```json
{"product_id": "1", "quantity": 1}
```

## Monitoring

The server prints stats every 5 seconds:
```
📊 STATS: Total=1500 | Success=450 (30.0%) | Failed=1050 | Timeouts=1045 | Queue=55 | MaxQueue=1200
```

## System Behavior

**What breaks during the flash sale?**
- ✅ System stays up
- ❌ Success rate drops dramatically (30% or less)
- ❌ Queue depth explodes (1000+ orders waiting)
- ❌ Customer experience degrades (30s timeouts)
- ❌ Your reputation 📉

**The bottleneck is realistic:** External payment processors have fixed throughput limits that can't be exceeded by simply adding more servers.

## Solution Strategies (Not Implemented)

To handle flash sales, you would need:
1. **Queue-based architecture** - Accept orders immediately, process async
2. **Circuit breaker pattern** - Fail fast when payment processor is overloaded
3. **Rate limiting** - Control incoming request rate
4. **Horizontal scaling** - More instances (but payment bottleneck remains!)
5. **Better UX** - Show queue position, estimated wait time

This demo shows why distributed systems need more than just "throw more servers at it"! 🚀

# problems
![alt text](image.png)
[::] means it's listening on IPv6, but your ECS tasks are using IPv4 addresses (10.0.10.31, etc.)!

```bash
e.Logger.Fatal(e.Start(":8080"))
#Change it to explicitly listen on IPv4:
e.Logger.Fatal(e.Start("0.0.0.0:8080"))
# Update your main.go with the change above
#$Rebuild and push your Docker image:

   docker build -t ecr_service:latest .
   
   # Tag for ECR
   docker tag ecr_service:latest 637423451078.dkr.ecr.us-west-2.amazonaws.com/ecr_service:latest
   
   # Push to ECR
   docker push 637423451078.dkr.ecr.us-west-2.amazonaws.com/ecr_service:latest
```
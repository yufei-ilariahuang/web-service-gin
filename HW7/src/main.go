package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/labstack/echo/v4"
)

// Product defines model for Product.
type Product struct {
	Brand       string  `json:"brand"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
}

// Item represents an item in an order
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float32 `json:"price"`
}

// Order defines the complete order structure
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending, processing, completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// OrderRequest defines the input for creating an order
type OrderRequest struct {
	CustomerID int    `json:"customer_id"`
	Items      []Item `json:"items"`
}

// OrderResponse defines the response for an order
type OrderResponse struct {
	OrderID        string    `json:"order_id"`
	CustomerID     int       `json:"customer_id"`
	Status         string    `json:"status"`
	Items          []Item    `json:"items"`
	TotalPrice     float32   `json:"total_price"`
	CreatedAt      time.Time `json:"created_at"`
	ProcessingTime string    `json:"processing_time"`
	QueueTime      string    `json:"queue_time,omitempty"`
}

// Stats tracks system performance
type Stats struct {
	TotalOrders      int64
	SuccessfulOrders int64
	FailedOrders     int64
	TimeoutOrders    int64
	CurrentQueue     int64
	MaxQueue         int64
	mu               sync.Mutex
}

var (
	products         []Product
	brands           = []string{"Apple", "Samsung", "Google", "Microsoft", "Sony", "Logitech"}
	categories       = []string{"Electronics", "Books", "Home", "Clothing", "Groceries", "Toys"}
	paymentSemaphore = make(chan int, 5) // Limits to 5 concurrent payment verifications
	orderCounter     int64
	stats            Stats
)

func main() {
	rand.Seed(time.Now().UnixNano())
	generateProducts(100000)

	e := echo.New()
	// Health check endpoint (for ALB)
	e.GET("/health", healthCheck)

	// Product routes
	e.GET("/products", getProducts)
	e.GET("/products/:id", getProductById)

	// Order routes - SYNCHRONOUS processing
	e.POST("/orders/sync", createOrderSync)
	e.GET("/stats", getStats)

	// Start stats printer
	go printStats()

	fmt.Println("🚀 Server starting on :8080")
	fmt.Println("📊 Synchronous order processing")
	fmt.Println("💳 Payment verification: 3 seconds per order")
	fmt.Println("⚡ Concurrent payment capacity: 5 orders")
	fmt.Println("🔥 Watch what breaks during flash sales!\n")

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

// healthCheck returns a simple 200 OK for ALB health checks
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "ecommerce-api",
	})
}

func generateProducts(count int) {
	products = make([]Product, count)
	for i := 0; i < count; i++ {
		brand := brands[rand.Intn(len(brands))]
		category := categories[rand.Intn(len(categories))]
		productName := fmt.Sprintf("Product %s %d", gofakeit.AppName(), i+1)

		products[i] = Product{
			ID:          strconv.Itoa(i + 1),
			Name:        productName,
			Category:    category,
			Brand:       brand,
			Description: gofakeit.BS(),
			Price:       rand.Float32() * 1000,
		}
	}
	fmt.Printf("✅ Generated %d products.\n\n", count)
}

func getProducts(c echo.Context) error {
	brandsParam := c.QueryParam("brands")
	categoriesParam := c.QueryParam("categories")

	var filteredProducts []Product
	limit := 100
	if len(products) < limit {
		limit = len(products)
	}

	for i := 0; i < limit; i++ {
		product := products[i]
		match := true

		if brandsParam != "" {
			if !strings.Contains(strings.ToLower(product.Brand), strings.ToLower(brandsParam)) {
				match = false
			}
		}

		if match && categoriesParam != "" {
			if !strings.EqualFold(product.Category, categoriesParam) {
				match = false
			}
		}

		if match {
			filteredProducts = append(filteredProducts, product)
		}
	}

	if len(filteredProducts) > 20 {
		filteredProducts = filteredProducts[:20]
	}

	return c.JSON(http.StatusOK, filteredProducts)
}

func getProductById(c echo.Context) error {
	id := c.Param("id")
	for _, p := range products {
		if p.ID == id {
			return c.JSON(http.StatusOK, p)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound, "Product not found")
}

// createOrderSync implements SYNCHRONOUS order processing with payment verification bottleneck
func createOrderSync(c echo.Context) error {
	startTime := time.Now()
	atomic.AddInt64(&stats.TotalOrders, 1)

	// Track queue depth
	currentQueue := atomic.AddInt64(&stats.CurrentQueue, 1)
	stats.mu.Lock()
	if currentQueue > stats.MaxQueue {
		stats.MaxQueue = currentQueue
	}
	stats.mu.Unlock()

	defer atomic.AddInt64(&stats.CurrentQueue, -1)

	// Parse request
	orderReq := new(OrderRequest)
	if err := c.Bind(orderReq); err != nil {
		atomic.AddInt64(&stats.FailedOrders, 1)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate customer ID
	if orderReq.CustomerID <= 0 {
		atomic.AddInt64(&stats.FailedOrders, 1)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer_id")
	}

	// Validate items
	if len(orderReq.Items) == 0 {
		atomic.AddInt64(&stats.FailedOrders, 1)
		return echo.NewHTTPError(http.StatusBadRequest, "Order must have at least one item")
	}

	// Validate all products exist and enrich items with prices
	var totalPrice float32
	validatedItems := make([]Item, len(orderReq.Items))
	for i, item := range orderReq.Items {
		var product *Product
		for j := range products {
			if products[j].ID == item.ProductID {
				product = &products[j]
				break
			}
		}

		if product == nil {
			atomic.AddInt64(&stats.FailedOrders, 1)
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Product %s not found", item.ProductID))
		}

		if item.Quantity <= 0 {
			atomic.AddInt64(&stats.FailedOrders, 1)
			return echo.NewHTTPError(http.StatusBadRequest, "Item quantity must be positive")
		}

		validatedItems[i] = Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
		totalPrice += product.Price * float32(item.Quantity)
	}

	// Generate order ID
	orderID := fmt.Sprintf("ORD-%d", atomic.AddInt64(&orderCounter, 1))
	createdAt := time.Now()
	queueTime := time.Since(startTime)

	// SYNCHRONOUS PAYMENT VERIFICATION with semaphore bottleneck
	// This simulates the real constraint: payment processor can only handle limited throughput
	select {
	case paymentSemaphore <- 1: // Acquire semaphore (blocks if 5 payments already processing)
		defer func() { <-paymentSemaphore }() // Release semaphore

		// Simulate 3-second payment verification delay
		// This represents the external payment processor bottleneck
		time.Sleep(3 * time.Second)

		atomic.AddInt64(&stats.SuccessfulOrders, 1)

		response := OrderResponse{
			OrderID:        orderID,
			CustomerID:     orderReq.CustomerID,
			Status:         "completed",
			Items:          validatedItems,
			TotalPrice:     totalPrice,
			CreatedAt:      createdAt,
			ProcessingTime: time.Since(startTime).Round(time.Millisecond).String(),
			QueueTime:      queueTime.Round(time.Millisecond).String(),
		}

		return c.JSON(http.StatusOK, response)

	case <-time.After(30 * time.Second): // Timeout after 30 seconds of waiting
		atomic.AddInt64(&stats.TimeoutOrders, 1)
		atomic.AddInt64(&stats.FailedOrders, 1)

		response := OrderResponse{
			OrderID:        orderID,
			CustomerID:     orderReq.CustomerID,
			Status:         "timeout",
			Items:          validatedItems,
			TotalPrice:     totalPrice,
			CreatedAt:      createdAt,
			ProcessingTime: time.Since(startTime).Round(time.Millisecond).String(),
			QueueTime:      queueTime.Round(time.Millisecond).String(),
		}

		return c.JSON(http.StatusServiceUnavailable, response)
	}
}

func getStats(c echo.Context) error {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	total := atomic.LoadInt64(&stats.TotalOrders)
	successful := atomic.LoadInt64(&stats.SuccessfulOrders)
	successRate := 0.0
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_orders":      total,
		"successful_orders": successful,
		"failed_orders":     atomic.LoadInt64(&stats.FailedOrders),
		"timeout_orders":    atomic.LoadInt64(&stats.TimeoutOrders),
		"current_queue":     atomic.LoadInt64(&stats.CurrentQueue),
		"max_queue_seen":    stats.MaxQueue,
		"success_rate":      fmt.Sprintf("%.2f%%", successRate),
	})
}

func printStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		total := atomic.LoadInt64(&stats.TotalOrders)
		if total == 0 {
			continue
		}

		successful := atomic.LoadInt64(&stats.SuccessfulOrders)
		failed := atomic.LoadInt64(&stats.FailedOrders)
		timeouts := atomic.LoadInt64(&stats.TimeoutOrders)
		queue := atomic.LoadInt64(&stats.CurrentQueue)

		stats.mu.Lock()
		maxQueue := stats.MaxQueue
		stats.mu.Unlock()

		successRate := float64(successful) / float64(total) * 100

		fmt.Printf("\n📊 STATS: Total=%d | Success=%d (%.1f%%) | Failed=%d | Timeouts=%d | Queue=%d | MaxQueue=%d\n",
			total, successful, successRate, failed, timeouts, queue, maxQueue)
	}
}

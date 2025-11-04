package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type TestResult struct {
	Operation    string    `json:"operation"`     // create_cart|add_items|get_cart
	ResponseTime float64   `json:"response_time"` // in milliseconds
	Success      bool      `json:"success"`
	StatusCode   int       `json:"status_code"`
	Timestamp    time.Time `json:"timestamp"`
}

type Cart struct {
	ShoppingCartId int32 `json:"shopping_cart_id"`
	CustomerId     int32 `json:"customer_id"`
}

type CartItem struct {
	ProductID int32 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

func runTest(baseURL string, outputFile string) error {
	var results []TestResult
	var cartIDs []int32 // Store created cart IDs as int32

	// 1. Create 50 carts
	fmt.Println("Creating 50 carts...")
	for i := 0; i < 50; i++ {
		start := time.Now()

		// Create cart payload
		payload := map[string]int32{
			"customer_id": int32(i + 1), // Adding 1 to ensure positive IDs
		}
		jsonData, _ := json.Marshal(payload)

		// Make request
		resp, err := http.Post(baseURL+"/shopping-carts", "application/json", bytes.NewBuffer(jsonData))
		end := time.Now()

		result := TestResult{
			Operation:    "create_cart",
			ResponseTime: float64(end.Sub(start).Microseconds()) / 1000.0, // Convert to milliseconds
			Timestamp:    start,
		}

		if err != nil {
			result.Success = false
			result.StatusCode = 500
		} else {
			result.Success = resp.StatusCode == 201 || resp.StatusCode == 200
			result.StatusCode = resp.StatusCode

			// Read cart ID from response if successful
			if result.Success {
				var cart Cart
				body, _ := io.ReadAll(resp.Body)
				json.Unmarshal(body, &cart)
				cartIDs = append(cartIDs, cart.ShoppingCartId)
				fmt.Printf("Created cart ID: %d\n", cart.ShoppingCartId)
			}
			resp.Body.Close()
		}

		results = append(results, result)
	}

	// 2. Add items to each cart
	fmt.Println("Adding items to carts...")
	for i, cartID := range cartIDs {
		start := time.Now()

		// Create item payload
		item := CartItem{
			ProductID: int32(i + 1),       // Product IDs from 1 onwards
			Quantity:  int32((i % 5) + 1), // Quantity between 1-5
		}
		jsonData, _ := json.Marshal(item)

		// Make request - correct URL with cartID
		url := fmt.Sprintf("%s/shopping-carts/%d/items", baseURL, cartID)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		end := time.Now()

		result := TestResult{
			Operation:    "add_items",
			ResponseTime: float64(end.Sub(start).Microseconds()) / 1000.0,
			Timestamp:    start,
		}

		if err != nil {
			result.Success = false
			result.StatusCode = 500
			fmt.Printf("Error adding item to cart %d: %v\n", cartID, err)
		} else {
			result.Success = resp.StatusCode == 204 || resp.StatusCode == 200
			result.StatusCode = resp.StatusCode
			if !result.Success {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("Add item failed for cart %d: status %d, body: %s\n", cartID, resp.StatusCode, string(body))
			}
			resp.Body.Close()
		}

		results = append(results, result)
	}

	// 3. Get each cart
	fmt.Println("Getting cart details...")
	for _, cartID := range cartIDs {
		start := time.Now()

		// Make request - correct URL with cartID
		url := fmt.Sprintf("%s/shopping-carts/%d", baseURL, cartID)
		resp, err := http.Get(url)
		end := time.Now()

		result := TestResult{
			Operation:    "get_cart",
			ResponseTime: float64(end.Sub(start).Microseconds()) / 1000.0,
			Timestamp:    start,
		}

		if err != nil {
			result.Success = false
			result.StatusCode = 500
			fmt.Printf("Error getting cart %d: %v\n", cartID, err)
		} else {
			result.Success = resp.StatusCode == 200
			result.StatusCode = resp.StatusCode
			if !result.Success {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("Get cart failed for cart %d: status %d, body: %s\n", cartID, resp.StatusCode, string(body))
			}
			resp.Body.Close()
		}

		results = append(results, result)
	}

	// Write results to file
	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling results: %v", err)
	}

	return os.WriteFile(outputFile, resultJSON, 0644)
}

func main() {
	baseURL := os.Getenv("API_URL")
	if baseURL == "" {
		baseURL = "http://CS6650L2-alb-397575304.us-west-2.elb.amazonaws.com" // use ALB endpoint
	}

	// Get backend from command line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run runner.go [mysql|dynamodb]")
		os.Exit(1)
	}

	backend := os.Args[1]
	if backend != "mysql" && backend != "dynamodb" {
		fmt.Println("Backend must be either 'mysql' or 'dynamodb'")
		os.Exit(1)
	}

	// Configure test based on backend
	config := struct {
		name       string
		outputFile string
		envVars    map[string]string
	}{
		name:       backend,
		outputFile: fmt.Sprintf("%s_test_results.json", backend),
		envVars: map[string]string{
			"DB_BACKEND": backend,
		},
	}

	fmt.Printf("\nTesting %s backend...\n", config.name)
	fmt.Println("This will run 150 operations (50 create, 50 add items, 50 get)")
	fmt.Println("Please start monitoring the following metrics in CloudWatch:")

	if backend == "mysql" {
		fmt.Println("MySQL Metrics to Monitor:")
		fmt.Println("- RDS CPU utilization")
		fmt.Println("- RDS connections")
		fmt.Println("- ECS task performance")
		fmt.Println("- Database I/O")
		fmt.Println("- Query performance")
	} else {
		fmt.Println("DynamoDB Metrics to Monitor:")
		fmt.Println("- Request latency")
		fmt.Println("- Consumed capacity")
		fmt.Println("- Throttling events")
		fmt.Println("- Partition distribution")
		fmt.Println("- Error rates")
	}

	fmt.Println("\nPress Enter to start the test...")
	fmt.Scanln() // Wait for user to press Enter

	// Set environment variables
	for k, v := range config.envVars {
		os.Setenv(k, v)
	}

	startTime := time.Now()
	if err := runTest(baseURL, config.outputFile); err != nil {
		fmt.Printf("Error running tests for %s: %v\n", config.name, err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nTests completed successfully for %s!\n", config.name)
	fmt.Printf("Total duration: %v\n", duration)
	fmt.Printf("Results written to %s\n", config.outputFile)
	fmt.Println("\nPlease collect the CloudWatch metrics now.")
}

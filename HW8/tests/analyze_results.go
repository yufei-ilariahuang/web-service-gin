package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

type TestResult struct {
	Operation    string    `json:"operation"`     // create_cart|add_items|get_cart
	ResponseTime float64   `json:"response_time"` // in milliseconds
	Success      bool      `json:"success"`
	StatusCode   int       `json:"status_code"`
	Timestamp    time.Time `json:"timestamp"`
}

type CombinedResults struct {
	MySQL    []TestResult `json:"mysql"`
	DynamoDB []TestResult `json:"dynamodb"`
}

type Metrics struct {
	AvgResponseTime float64
	P50ResponseTime float64
	P95ResponseTime float64
	P99ResponseTime float64
	SuccessRate     float64
	TotalOps        int
}

type OperationMetrics struct {
	Operation string
	AvgTime   float64
}

func calculatePercentile(times []float64, percentile float64) float64 {
	if len(times) == 0 {
		return 0
	}

	sort.Float64s(times)
	index := int(math.Ceil((percentile / 100) * float64(len(times))))
	if index >= len(times) {
		index = len(times) - 1
	}
	return times[index]
}

func calculateMetrics(results []TestResult) (Metrics, map[string]float64) {
	if len(results) == 0 {
		return Metrics{}, nil
	}

	var responseTimes []float64
	successCount := 0
	operationTimes := make(map[string][]float64)

	for _, result := range results {
		responseTimes = append(responseTimes, result.ResponseTime)
		if result.Success {
			successCount++
		}
		operationTimes[result.Operation] = append(operationTimes[result.Operation], result.ResponseTime)
	}

	// Calculate average response times per operation
	operationAvgs := make(map[string]float64)
	for op, times := range operationTimes {
		sum := 0.0
		for _, t := range times {
			sum += t
		}
		operationAvgs[op] = sum / float64(len(times))
	}

	// Calculate overall metrics
	sum := 0.0
	for _, time := range responseTimes {
		sum += time
	}

	return Metrics{
		AvgResponseTime: sum / float64(len(responseTimes)),
		P50ResponseTime: calculatePercentile(responseTimes, 50),
		P95ResponseTime: calculatePercentile(responseTimes, 95),
		P99ResponseTime: calculatePercentile(responseTimes, 99),
		SuccessRate:     (float64(successCount) / float64(len(results))) * 100,
		TotalOps:        len(results),
	}, operationAvgs
}

func main() {
	// Read MySQL results
	mysqlData, err := os.ReadFile("mysql_test_results.json")
	if err != nil {
		fmt.Printf("Error reading MySQL results: %v\n", err)
		return
	}

	// Read DynamoDB results
	dynamoData, err := os.ReadFile("dynamodb_test_results.json")
	if err != nil {
		fmt.Printf("Error reading DynamoDB results: %v\n", err)
		return
	}

	var mysqlResults, dynamoResults []TestResult
	if err := json.Unmarshal(mysqlData, &mysqlResults); err != nil {
		fmt.Printf("Error parsing MySQL results: %v\n", err)
		return
	}
	if err := json.Unmarshal(dynamoData, &dynamoResults); err != nil {
		fmt.Printf("Error parsing DynamoDB results: %v\n", err)
		return
	}

	// Save combined results
	combined := CombinedResults{
		MySQL:    mysqlResults,
		DynamoDB: dynamoResults,
	}
	combinedJSON, _ := json.MarshalIndent(combined, "", "  ")
	os.WriteFile("combined_results.json", combinedJSON, 0644)

	// Calculate metrics
	mysqlMetrics, mysqlOpAvgs := calculateMetrics(mysqlResults)
	dynamoMetrics, dynamoOpAvgs := calculateMetrics(dynamoResults)

	// Print Performance Comparison Table
	fmt.Println("\nPerformance Comparison Table")
	fmt.Println("============================")
	fmt.Printf("Metric               MySQL        DynamoDB     Winner        Margin\n")
	fmt.Printf("------------------------------------------------------------------\n")
	fmt.Printf("Avg Response Time    %.2f ms     %.2f ms     %s          %.2f ms\n",
		mysqlMetrics.AvgResponseTime,
		dynamoMetrics.AvgResponseTime,
		winner(mysqlMetrics.AvgResponseTime, dynamoMetrics.AvgResponseTime),
		math.Abs(mysqlMetrics.AvgResponseTime-dynamoMetrics.AvgResponseTime))

	fmt.Printf("P50 Response Time    %.2f ms     %.2f ms     %s          %.2f ms\n",
		mysqlMetrics.P50ResponseTime,
		dynamoMetrics.P50ResponseTime,
		winner(mysqlMetrics.P50ResponseTime, dynamoMetrics.P50ResponseTime),
		math.Abs(mysqlMetrics.P50ResponseTime-dynamoMetrics.P50ResponseTime))

	fmt.Printf("P95 Response Time    %.2f ms     %.2f ms     %s          %.2f ms\n",
		mysqlMetrics.P95ResponseTime,
		dynamoMetrics.P95ResponseTime,
		winner(mysqlMetrics.P95ResponseTime, dynamoMetrics.P95ResponseTime),
		math.Abs(mysqlMetrics.P95ResponseTime-dynamoMetrics.P95ResponseTime))

	fmt.Printf("P99 Response Time    %.2f ms     %.2f ms     %s          %.2f ms\n",
		mysqlMetrics.P99ResponseTime,
		dynamoMetrics.P99ResponseTime,
		winner(mysqlMetrics.P99ResponseTime, dynamoMetrics.P99ResponseTime),
		math.Abs(mysqlMetrics.P99ResponseTime-dynamoMetrics.P99ResponseTime))

	fmt.Printf("Success Rate         %.2f%%      %.2f%%      %s          %.2f%%\n",
		mysqlMetrics.SuccessRate,
		dynamoMetrics.SuccessRate,
		winner(mysqlMetrics.SuccessRate, dynamoMetrics.SuccessRate),
		math.Abs(mysqlMetrics.SuccessRate-dynamoMetrics.SuccessRate))

	fmt.Printf("Total Operations     %d          %d          -             -\n",
		mysqlMetrics.TotalOps,
		dynamoMetrics.TotalOps)

	// Print Operation-Specific Breakdown
	fmt.Println("\nOperation-Specific Breakdown")
	fmt.Println("===========================")
	fmt.Printf("Operation       MySQL Avg (ms)    DynamoDB Avg (ms)    Faster By\n")
	fmt.Printf("----------------------------------------------------------------\n")

	ops := []string{"create_cart", "add_items", "get_cart"}
	for _, op := range ops {
		mysqlAvg := mysqlOpAvgs[op]
		dynamoAvg := dynamoOpAvgs[op]
		faster := ""
		if mysqlAvg < dynamoAvg {
			faster = "MySQL by %.2f ms"
			faster = fmt.Sprintf(faster, dynamoAvg-mysqlAvg)
		} else {
			faster = "DynamoDB by %.2f ms"
			faster = fmt.Sprintf(faster, mysqlAvg-dynamoAvg)
		}
		fmt.Printf("%-14s %-16.2f %-18.2f %s\n", op, mysqlAvg, dynamoAvg, faster)
	}
}

func winner(mysql, dynamo float64) string {
	if mysql < dynamo {
		return "MySQL    "
	}
	return "DynamoDB "
}

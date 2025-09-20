package main

import (
	"fmt"
	"sync"
	"time"
)

// Container uses a standard sync.Mutex.
// Every operation, read or write, requires an exclusive lock.
type Container struct {
	mu       sync.Mutex
	counters map[string]int
}

// inc is the writer method.
func (c *Container) inc(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name]++
}

// get is the reader method.
func (c *Container) get(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counters[name]
}

// --- Benchmark 1: WRITE-ONLY ---
// This function measures the performance of only the write operations.
func runWriteOnlyBenchmark() {
	c := Container{counters: map[string]int{"a": 0, "b": 0}}
	var wg sync.WaitGroup

	doIncrement := func(name string, n int) {
		defer wg.Done()
		for i := 0; i < n; i++ {
			c.inc(name)
		}
	}

	// Start timer for the write test
	startTime := time.Now()

	// Launch 3 writer goroutines
	wg.Add(3)
	go doIncrement("a", 10000)
	go doIncrement("a", 10000)
	go doIncrement("b", 10000)

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Println("--- 1. Write-Only Benchmark ---")
	fmt.Printf("Execution time (3 writers): %v\n", duration)
	fmt.Printf("Final State: a=%d, b=%d\n\n", c.counters["a"], c.counters["b"])
}

// --- Benchmark 2: READ-ONLY ---
// This function measures the performance of only the read operations.
func runReadOnlyBenchmark() {
	// Pre-populate the map with data to ensure reads have values.
	c := Container{counters: map[string]int{"a": 20000, "b": 10000}}
	var wg sync.WaitGroup

	doRead := func(name string, n int) {
		defer wg.Done()
		for i := 0; i < n; i++ {
			_ = c.get(name)
		}
	}

	// Start timer for the read test
	startTime := time.Now()

	// Launch 10 reader goroutines
	const numReaders = 10
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go doRead("a", 10000)
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Println("--- 2. Read-Only Benchmark ---")
	fmt.Printf("Execution time (10 readers): %v\n", duration)
}

func Mutexes() {
	fmt.Println("Running sequential benchmarks for Mutex-based container...")
	fmt.Println("=========================================================")

	// Call the function to run the write-only benchmark
	runWriteOnlyBenchmark()

	// Call the function to run the read-only benchmark
	runReadOnlyBenchmark()

	fmt.Println("=========================================================")
	fmt.Println("Benchmarks complete.")
}

// The main entry point of the program.

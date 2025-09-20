package main

import (
	"fmt"
	"sync"
	"time"
)

// --- Benchmark 1: WRITE-ONLY for sync.Map ---
// This function measures the performance of concurrent, low-contention writes.
func runSyncMapWriteOnlyBenchmark() {
	var m sync.Map
	var wg sync.WaitGroup

	const numWriters = 3
	const numWritesPerGoroutine = 1000

	// Start timer for the write test
	startTime := time.Now()

	// Launch 3 writer goroutines
	wg.Add(numWriters)
	for g := 0; g < numWriters; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			// Each goroutine writes to a unique range of keys
			// to simulate a low-contention workload.
			for i := 0; i < numWritesPerGoroutine; i++ {
				// Store is the concurrent-safe way to write.
				m.Store(goroutineID*numWritesPerGoroutine+i, i)
			}
		}(g)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Count entries to verify correctness
	var count int
	m.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	fmt.Println("--- 1. sync.Map Write-Only Benchmark ---")
	fmt.Printf("Execution time (3 writers, %d total writes): %v\n", count, duration)
	fmt.Printf("Final State: %d entries\n\n", count)
}

// --- Benchmark 2: READ-ONLY for sync.Map ---
// This function measures the performance of concurrent reads.
func runSyncMapReadOnlyBenchmark() {
	var m sync.Map
	var wg sync.WaitGroup

	// Step 1: Pre-populate the map with data to read.
	// We need existing keys to test the read performance.
	const totalKeys = 3000
	for i := 0; i < totalKeys; i++ {
		m.Store(i, i)
	}

	doRead := func(n int) {
		defer wg.Done()
		for i := 0; i < n; i++ {
			// Load is the concurrent-safe way to read.
			// We read key 0 repeatedly to ensure the key is always present.
			// This tests the "fast path" for reads in sync.Map.
			_, _ = m.Load(0)
		}
	}

	// Start timer for the read test
	startTime := time.Now()

	// Launch 10 reader goroutines, simulating a read-heavy scenario
	const numReaders = 10
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go doRead(10000)
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Println("--- 2. sync.Map Read-Only Benchmark ---")
	fmt.Printf("Execution time (10 readers, 100,000 total reads): %v\n", duration)
}

// SyncMapExperiment is the main function that orchestrates the benchmarks.
func SyncMapExperiment() {
	fmt.Println("Running sequential benchmarks for sync.Map container...")
	fmt.Println("======================================================")

	// Call the function to run the write-only benchmark
	runSyncMapWriteOnlyBenchmark()

	// Call the function to run the read-only benchmark
	runSyncMapReadOnlyBenchmark()

	fmt.Println("======================================================")
	fmt.Println("Benchmarks complete.")
}

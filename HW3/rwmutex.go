package main

import (
	"fmt"
	"sync"
	"time"
)

type Container2 struct {
	// mu is a Reader/Writer mutual exclusion lock.
	// It allows many concurrent readers OR one single writer.
	mu       sync.RWMutex
	counters map[string]int
}

// inc2 is the "writer" method. It must acquire an exclusive lock.
func (c *Container2) inc2(name string) {
	// c.mu.Lock() prevents any other goroutine (reader or writer) from
	// proceeding until this one is done.
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name]++
}

// get is the "reader" method. It can acquire a shared lock.
func (c *Container2) get(name string) int {
	// c.mu.RLock() allows multiple goroutines to acquire this lock
	// simultaneously, as long as no writer holds the exclusive lock.
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.counters[name]
}

// --- Benchmark 1: WRITE-ONLY for RWMutex ---
// This measures the performance of high-contention writes.
func runRWMutexWriteOnlyBenchmark() {
	c := Container2{counters: map[string]int{"a": 0, "b": 0}}
	var wg sync.WaitGroup

	doIncrement := func(name string, n int) {
		defer wg.Done()
		for i := 0; i < n; i++ {
			c.inc2(name)
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

	fmt.Println("--- 1. RWMutex Write-Only Benchmark ---")
	fmt.Printf("Execution time (3 writers): %v\n", duration)
	fmt.Printf("Final State: a=%d, b=%d\n\n", c.counters["a"], c.counters["b"])
}

// --- Benchmark 2: READ-ONLY for RWMutex ---
// This measures the performance of concurrent reads.
func runRWMutexReadOnlyBenchmark() {
	// Pre-populate the map with data for the readers.
	c := Container2{counters: map[string]int{"a": 20000, "b": 10000}}
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

	fmt.Println("--- 2. RWMutex Read-Only Benchmark ---")
	fmt.Printf("Execution time (10 readers): %v\n", duration)
}

// RWMutexes is the main function that orchestrates the benchmarks.
func RWMutexes() {
	fmt.Println("Running sequential benchmarks for RWMutex-based container...")
	fmt.Println("==========================================================")

	runRWMutexWriteOnlyBenchmark()
	runRWMutexReadOnlyBenchmark()

	fmt.Println("==========================================================")
	fmt.Println("Benchmarks complete.")
}

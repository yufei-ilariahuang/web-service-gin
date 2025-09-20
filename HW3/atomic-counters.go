package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func AtomicCounters() {
	// 1. The Shared Counter (The Whiteboard)
	// We declare our counter using a special type from the atomic package.
	// This type has special methods like .Add() and .Load()
	var ops atomic.Uint64

	// 2. The Coordinator
	// A WaitGroup is used to make the main function wait until all its
	// child goroutines have finished their work.
	var wg sync.WaitGroup

	// 3. Starting the Workers
	// We start 50 goroutines. Each one is a separate "worker".

	for range 50 {

		// 'go func()' starts a new goroutine.
		wg.Go(func() {
			for range 1000 {
				// 4. The Atomic Add()
				// The safe way to increment the counter.
				// It prevents the race condition.
				ops.Add(1)
			}
		})
	}

	// 5. Waiting for Everyone to Finish
	// wg.Wait() pauses the main function here. It will not continue
	// until wg.Done() has been called 50 times (once by each worker).
	wg.Wait()

	// 6. Reading the Final Result
	// We are now sure all 50,000 increments are done.
	// ops.Load() is the safe, atomic way to read the value, ensuring
	// we get the complete final number without any corruption.
	fmt.Println("ops:", ops.Load())
}

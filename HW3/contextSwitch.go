package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// runPingPongTest executes the core logic of the context switching benchmark.
// It launches two goroutines that pass a signal back and forth `numSwitches` times
// and returns the total duration of the experiment.
func runPingPongTest(numSwitches int) time.Duration {
	// An unbuffered channel ensures that each send/receive is a synchronization
	// point, forcing a context switch between the two goroutines.
	ch := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	startTime := time.Now()

	// Goroutine 1: Sends signals and then waits for the final signal.
	go func() {
		defer wg.Done()
		for i := 0; i < numSwitches; i++ {
			ch <- struct{}{}
		}
	}()

	// Goroutine 2: Receives signals.
	go func() {
		defer wg.Done()
		for i := 0; i < numSwitches; i++ {
			<-ch
		}
	}()

	// Wait for both goroutines to complete their work.
	wg.Wait()

	// Return the total time elapsed.
	return time.Since(startTime)
}

// runAllExperiments is the single function that executes and prints the results
// for both the single-threaded and multi-threaded scenarios.
func contextSwitch() {
	const numRoundTrips = 1000000
	// Total switches is 2 * numRoundTrips (one for send, one for receive).
	const totalSwitches = 2 * numRoundTrips

	fmt.Println("--- Context Switching Benchmark ---")
	fmt.Printf("Performing %d round-trips (%d total context switches).\n\n", numRoundTrips, totalSwitches)

	// --- Scenario 1: Single OS Thread ---
	// Restrict the Go runtime to a single operating system thread.
	runtime.GOMAXPROCS(1)
	durationSingleThread := runPingPongTest(numRoundTrips)
	avgSwitchTimeSingle := durationSingleThread / totalSwitches

	fmt.Printf("Scenario 1: Single OS Thread (GOMAXPROCS=1)\n")
	fmt.Printf("Total duration: %v\n", durationSingleThread)
	fmt.Printf("Average context switch time: %v\n\n", avgSwitchTimeSingle)

	// --- Scenario 2: Multiple OS Threads ---
	// Allow the Go runtime to use all available CPU cores.
	// runtime.NumCPU() returns the number of logical CPUs usable by the current process.
	numCores := runtime.NumCPU()
	runtime.GOMAXPROCS(numCores)
	durationMultiThread := runPingPongTest(numRoundTrips)
	avgSwitchTimeMulti := durationMultiThread / totalSwitches

	fmt.Printf("Scenario 2: Multiple OS Threads (GOMAXPROCS=%d)\n", numCores)
	fmt.Printf("Total duration: %v\n", durationMultiThread)
	fmt.Printf("Average context switch time: %v\n\n", avgSwitchTimeMulti)

	// --- Conclusion ---
	fmt.Println("--- Comparison ---")
	if avgSwitchTimeSingle < avgSwitchTimeMulti {
		fmt.Printf("The single-threaded version was faster.\n")
		fmt.Println("This is because the Go runtime can schedule goroutine switches in user space without the overhead of kernel-level thread synchronization across multiple CPU cores.")
	} else {
		fmt.Printf("The multi-threaded version was faster.\n")
		fmt.Println("This result might occur on some systems or Go versions, but typically the single-threaded version excels in this specific communication-heavy, non-parallelizable benchmark.")
	}
}

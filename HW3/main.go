// File: main.go
package main

import (
	"flag"
	"fmt"
)

func main() {
	// Define the command-line flag to choose which example to run.
	exampleToRun := flag.String("run", "all", "Specify which example to run: 'atomic', 'mutex', 'collections', 'RWMutexes', 'syncMap', 'fileAccess', 'contextSwitch', or 'all'")
	flag.Parse()

	// Decide which function to call based on the flag's value.
	switch *exampleToRun {
	case "atomic":
		AtomicCounters()
	case "mutex":
		Mutexes()
	case "collections":
		Collections()
	case "RWMutexes":
		RWMutexes()
	case "syncMap":
		SyncMapExperiment()
	case "fileAccess":
		fileAccess()
	case "contextSwitch":
		contextSwitch()
	case "all":
		AtomicCounters()
		Mutexes()
		Collections()
		RWMutexes()
		SyncMapExperiment()
		fileAccess()
		contextSwitch()
	default:
		fmt.Printf("Invalid example specified: %s\n", *exampleToRun)
		fmt.Println("Please use 'atomic', 'mutex', or 'all'.")
	}
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

const (
	iterations = 100000
	fileName   = "output.txt"
	line       = "This is a line of text.\n"
)

// unbufferedWrite writes to a file on each iteration.
func unbufferedWrite() time.Duration {
	// Open file for writing, create if it doesn't exist, truncate if it does.
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := f.Write([]byte(line))
		if err != nil {
			panic(err)
		}
	}

	return time.Since(start)
}

// bufferedWrite writes to a buffer, then flushes to the file.
func bufferedWrite() time.Duration {
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Wrap the file writer in a bufio.Writer.
	writer := bufio.NewWriter(f)

	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := writer.WriteString(line)
		if err != nil {
			panic(err)
		}
	}

	// Flush writes the buffered data to the underlying file.
	if err := writer.Flush(); err != nil {
		panic(err)
	}

	return time.Since(start)
}
func fileAccess() {
	unbufferedDuration := unbufferedWrite()
	fmt.Printf("Unbuffered write took: %v\n", unbufferedDuration)

	bufferedDuration := bufferedWrite()
	fmt.Printf("Buffered write took:   %v\n", bufferedDuration)
}

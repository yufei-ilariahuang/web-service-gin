package main

import (
	"fmt"
)

// counters is the map that stores our data.

func Collections() {
	var m = make(map[int]int)

	for g := 0; g < 50; g++ {

		go func(g int) {
			for i := 0; i < 1000; i++ {
				m[g*1000+i] = i
			}
		}(g)
	}
	fmt.Println(len(m))
}

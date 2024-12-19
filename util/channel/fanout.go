package channel

import (
	"sync"
)

// Transform applies a function to values from an input channel in parallel using goroutines
// and sends the transformed values to an output channel.
func Transform[K any, F any](procs int, input <-chan K, trans func(K) F) <-chan F {
	output := make(chan F, len(input)) // buffer the input the same as the output

	var wg sync.WaitGroup
	// Launch specified number of goroutines
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for val := range input {
				output <- trans(val)
			}
		}()
	}

	// Close the output channel when all workers finish
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

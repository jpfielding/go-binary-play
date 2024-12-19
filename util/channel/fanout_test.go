package channel_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/jpfielding/go-binary-play/util/channel"
)

func TestTransform(t *testing.T) {
	// Example usage
	input := make(chan int)

	// get the value without setting
	procs := runtime.GOMAXPROCS(0)
	output := channel.Transform(procs, input, func(n int) string {
		return fmt.Sprintf("Number: %d", n)
	})

	// Send some data into the input channel and close it
	go func() {
		for i := 0; i < 10; i++ {
			input <- i
		}
		close(input)
	}()

	// Read from the output channel
	for result := range output {
		fmt.Println(result)
	}
}

package channel_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jpfielding/go-binary-play/util/channel"
	"github.com/stretchr/testify/assert"
)

func TestTap(t *testing.T) {
	in := make(chan string, 1)
	ctx := context.Background()
	defer close(in)
	var found []string
	out := channel.Tap(ctx, in, func(s string) {
		found = append(found, s)
	})
	in <- "beep"
	<-out
	assert.Equal(t, 1, len(found))
	assert.Equal(t, "beep", found[0])
}

func TestTapLoad(t *testing.T) {
	in := make(chan string, 1)
	ctx := context.Background()
	defer close(in)
	var found []string
	out := channel.Tap(ctx, in, func(s string) {
		found = append(found, s)
	})
	var wg sync.WaitGroup
	each := 10000
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < each; i++ {
			in <- fmt.Sprintf("beep-%d", i)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < each; i++ {
			found := <-out
			assert.Equal(t, found, fmt.Sprintf("beep-%d", i))
		}
	}()
	wg.Wait()
	assert.Equal(t, each, len(found))
}

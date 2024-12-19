package channel_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jpfielding/go-binary-play/util/channel"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	in := make(chan string, 1)
	ctx := context.Background()
	defer close(in)
	out := channel.Filter(ctx, in, func(test string) bool {
		return strings.Contains(test, "bo")
	})
	t.Run("keep", func(t *testing.T) {
		in <- "boop"
		time.Sleep(time.Second)
		assert.Equal(t, "boop", <-out)
	})
	t.Run("drop", func(t *testing.T) {
		in <- "beep"
		select {
		case <-out:
			assert.Fail(t, "should have been dropped")
		default:
		}
	})
}

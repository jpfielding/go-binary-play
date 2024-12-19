package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/jpfielding/go-binary-play/util/channel"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	s := make(chan string, 1)
	ctx := context.Background()
	defer close(s)
	t.Run("still there", func(t *testing.T) {
		st := channel.Timeout(ctx, s, time.Second)
		s <- "beep"
		assert.Equal(t, "beep", <-st)
	})
	t.Run("time out", func(t *testing.T) {
		st := channel.Timeout(ctx, s, time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		select {
		case <-st:
			assert.Fail(t, "should have timed out")
		default:
		}
	})
}

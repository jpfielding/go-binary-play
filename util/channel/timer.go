package channel

import (
	"context"
	"time"
)

// Timeout discards the head element after ttl sat at head for ttl
func Timeout[A any](ctx context.Context, ch chan A, ttl time.Duration) chan A {
	tmp := make(chan A)
	//pull one out of our incoming ch, and either push it into our unbuffered chan before ttl or discard
	go func() {
		defer close(tmp)
		for c := range ch {
			select {
			case tmp <- c: // try to send to tmp, give up after 55l
			case <-ctx.Done():
				return
			case <-time.After(ttl):
			}
		}
	}()
	return tmp
}

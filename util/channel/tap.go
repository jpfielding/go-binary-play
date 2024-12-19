package channel

import "context"

// Tap a chan to listen for each value
func Tap[T any](ctx context.Context, in chan T, f func(t T)) chan T {
	tmp := make(chan T, cap(in))
	go func() {
		defer close(tmp)
		for {
			select {
			case <-ctx.Done():
				return
			case v := <-in:
				f(v)
				select {
				case <-ctx.Done():
					return
				case tmp <- v:
				}
			}
		}
	}()
	return tmp
}

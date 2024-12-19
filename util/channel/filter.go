package channel

import "context"

// Filter a chan to listen for each value
func Filter[T any](ctx context.Context, in chan T, keep func(t T) bool) chan T {
	tmp := make(chan T, cap(in))
	go func() {
		defer close(tmp)
		for {
			select {
			case <-ctx.Done():
				return
			case v := <-in:
				if !keep(v) {
					continue
				}
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

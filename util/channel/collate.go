package channel

import (
	"context"
	"sync"
)

type Collater[T any] struct {
	build   func(string) chan T
	collate map[string]chan T
	mut     sync.Mutex
}

func NewCollater[T any](build func(string) chan T) *Collater[T] {
	return &Collater[T]{
		build:   build,
		collate: map[string]chan T{},
	}
}

// Collate should be run in a goroutine. It will block until the context is cancelled or an OperatorDecision is received.
func (cc *Collater[T]) Collate(ctx context.Context, ch <-chan T, keyer func(t T) string) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ch:
			cc.Get(keyer(t)) <- t
		}
	}
}

func (cc *Collater[T]) Get(key string) chan T {
	cc.mut.Lock()
	defer cc.mut.Unlock()
	if _, ok := cc.collate[key]; !ok {
		cc.collate[key] = cc.build(key)
	}
	return cc.collate[key]
}

func (cc *Collater[T]) Count() map[string]int {
	cc.mut.Lock()
	defer cc.mut.Unlock()
	counts := map[string]int{}
	for k, v := range cc.collate {
		counts[k] = len(v)
	}
	return counts
}

func (cc *Collater[T]) Close() error {
	cc.mut.Lock()
	defer cc.mut.Unlock()
	for _, ch := range cc.collate {
		close(ch)
	}
	return nil
}

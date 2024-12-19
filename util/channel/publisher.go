package channel

import (
	"sync"
)

type PubSub[T any] struct {
	mu          sync.Mutex
	notReady    chan struct{}
	subscribers map[chan T]func()
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		notReady:    make(chan struct{}),
		subscribers: make(map[chan T]func()),
	}
}

func (ps *PubSub[T]) UntilReady() <-chan struct{} {
	return ps.notReady
}

func (ps *PubSub[T]) Subscribe(buffer int) (<-chan T, func()) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if len(ps.subscribers) == 0 {
		close(ps.notReady)
	}
	ch := make(chan T, buffer)
	// Unsubscribe function to remove the subscriber
	unsubscribe := func() {
		ps.mu.Lock()
		defer ps.mu.Unlock()
		delete(ps.subscribers, ch)
		close(ch)
		if len(ps.subscribers) == 0 {
			ps.notReady = make(chan struct{})
		}
	}
	ps.subscribers[ch] = unsubscribe

	return ch, unsubscribe
}

// Publish sends a message to all unblocked subscribers.
// If async is true, blocked cases are sent in the background.
func (ps *PubSub[T]) Publish(msg T, async bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for ch := range ps.subscribers {
		select {
		case ch <- msg:
			// Successfully sent the message
		default:
			if async {
				// Handle blocked case in the background
				go func(ch chan T) { ch <- msg }(ch)
			}
		}
	}
}

func (ps *PubSub[T]) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, unsub := range ps.subscribers {
		unsub()
	}
}

package channel_test

import (
	"testing"

	"github.com/jpfielding/go-binary-play/util/channel"

	"github.com/stretchr/testify/assert"
)

func TestNewPubSub(t *testing.T) {
	ps := channel.NewPubSub[string]()
	assert.NotNil(t, ps)
	s, us := ps.Subscribe(1)
	defer us()
	ps.Publish("test message", false)
	assert.Equal(t, "test message", <-s)
}

func TestPubSubReady(t *testing.T) {
	ps := channel.NewPubSub[string]()
	assert.NotNil(t, ps)
	select {
	case <-ps.UntilReady():
		assert.Fail(t, "PubSub should not be ready")
	default:
	}
	ps.Publish("eat message", false)
	s, us := ps.Subscribe(0)
	defer us()
	select {
	case <-ps.UntilReady():
	default:
		assert.Fail(t, "PubSub should not be ready")
	}
	ps.Publish("test message", false)
	select {
	case <-s:
		assert.Fail(t, "Sub should be dropped as read was not waiting")
	default:
	}
	ps.Publish("test message", true)
	assert.Equal(t, "test message", <-s)
}

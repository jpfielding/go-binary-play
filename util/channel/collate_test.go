package channel_test

import (
	"context"
	"testing"

	"github.com/jpfielding/go-binary-play/util/channel"
	"github.com/stretchr/testify/assert"
)

func TestCollator(t *testing.T) {
	ctx, cnc := context.WithCancel(context.Background())
	defer cnc()

	type test struct {
		ID string
	}

	tests := make(chan test, 1)
	defer close(tests)
	testers := channel.NewCollater[test](func(string) chan test {
		return make(chan test, cap(tests))
	})
	go testers.Collate(ctx, tests, func(t test) string { return t.ID })
	defer testers.Close()

	tests <- test{ID: "three"}
	tests <- test{ID: "one"}
	tests <- test{ID: "four"}
	tests <- test{ID: "two"}
	assert.Equal(t, "four", (<-testers.Get("four")).ID)
	assert.Equal(t, "three", (<-testers.Get("three")).ID)
	assert.Equal(t, "two", (<-testers.Get("two")).ID)
	assert.Equal(t, "one", (<-testers.Get("one")).ID)
	select {
	case <-testers.Get("none"):
		assert.Fail(t, "should not have received a decision for resource 'none'")
	default:
	}
}

package compute

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcessMustCallAllNodesInOrder(t *testing.T) {
	out := make(chan interface{}, 0)
	chainProcessor := NewChainProcessor(10,
		newNode(1),
		newNode(2),
		newReportingNode(3, out))

	chainProcessor.Start()
	defer chainProcessor.Shutdown()
	chainProcessor.Process(1)

	procOutput := <-out

	assert.Equal(t, 4, procOutput)
}

func TestProcessMustRejectWhenProcessingQueuesAreFull(t *testing.T) {
	chainProcessor := NewChainProcessor(1, newHeavyNode(newNode(1)))

	chainProcessor.Start()
	defer chainProcessor.Shutdown()
	chainProcessor.Process(1)
	chainProcessor.Process(1)

	assert.False(t, chainProcessor.Process(1))
}

func TestProcessMustProcessAllQueuedItemsBeforeShuttingDown(t *testing.T) {
	out := make(chan interface{}, 3)
	chainProcessor := NewChainProcessor(3, newHeavyNode(newReportingNode(1, out)))

	chainProcessor.Start()

	assert.True(t, chainProcessor.Process(1))
	assert.True(t, chainProcessor.Process(1))
	assert.True(t, chainProcessor.Process(1))

	chainProcessor.Shutdown()

	assert.Equal(t, 2, <-out)
	assert.Equal(t, 2, <-out)
	assert.Equal(t, 2, <-out)
}

func newNode(expectedPayload int) func(interface{}) interface{} {
	return func(p interface{}) interface{} {
		if p == expectedPayload {
			return expectedPayload + 1
		} else {
			return p
		}
	}
}

func newHeavyNode(delegate func(interface{}) interface{}) func(interface{}) interface{} {
	return func(p interface{}) interface{} {
		time.Sleep(time.Second * 2)
		return delegate(p)
	}
}

func newReportingNode(expectedPayload int, out chan interface{}) func(interface{}) interface{} {
	return func(p interface{}) interface{} {
		result := newNode(expectedPayload)(p)
		out <- result

		return result
	}
}

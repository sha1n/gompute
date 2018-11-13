package gompute

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcessMustCallAllNodesInOrder(t *testing.T) {
	out := make(chan interface{}, 0)
	chainProcessor, _ := NewChainProcessorBuilder().
		WithMaxQueueLength(10).
		AppendProcessNodes(newNode(1), newNode(2)).
		AppendProcessNodes(newReportingNode(3, out)).
		Build()

	chainProcessor.Start()
	defer chainProcessor.Shutdown()
	chainProcessor.Process(1)

	procOutput := <-out

	assert.Equal(t, 4, procOutput)
}

func TestProcessMustRejectItemsIfNotStarted(t *testing.T) {
	chainProcessor, _ := NewChainProcessorBuilder().
		WithMaxQueueLength(10).
		AppendProcessNodes(newNode(1)).
		Build()

	assert.False(t, chainProcessor.Process(1))
}

func TestProcessMustRejectItemsAfterShutdown(t *testing.T) {
	chainProcessor, _ := NewChainProcessorBuilder().
		WithMaxQueueLength(10).
		AppendProcessNodes(newNode(1)).
		Build()

	chainProcessor.Start()
	chainProcessor.Shutdown()
	assert.False(t, chainProcessor.Process(1))
}

func TestProcessMustRejectWhenProcessingQueuesAreFull(t *testing.T) {
	chainProcessor, _ := NewChainProcessorBuilder().
		WithMaxWorkers(1).
		WithMaxQueueLength(1).
		AppendProcessNodes(newHeavyNode(newNode(1))).
		Build()

	chainProcessor.Start()
	defer chainProcessor.Shutdown()
	chainProcessor.Process(1)
	chainProcessor.Process(1)

	assert.False(t, chainProcessor.Process(1))
}

func TestProcessMustProcessAllQueuedItemsBeforeShuttingDown(t *testing.T) {
	out := make(chan interface{}, 3)
	chainProcessor, _ := NewChainProcessorBuilder().
		WithMaxQueueLength(3).
		AppendProcessNodes(newHeavyNode(newReportingNode(1, out))).
		Build()

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

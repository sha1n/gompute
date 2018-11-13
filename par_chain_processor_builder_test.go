package gompute

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuilderMustValidateMaxWorkersValue(t *testing.T) {
	chainProcessor, err := NewChainProcessorBuilder().
		WithMaxWorkers(0).
		WithMaxQueueLength(10).
		AppendProcessNodes(newNode(1)).
		Build()

	assert.Error(t, err)
	assert.Nil(t, chainProcessor)
}

func TestBuilderMustValidateMaxQueueLengthValue(t *testing.T) {
	chainProcessor, err := NewChainProcessorBuilder().
		WithMaxQueueLength(0).
		AppendProcessNodes(newNode(1)).
		Build()

	assert.Error(t, err)
	assert.Nil(t, chainProcessor)
}

func TestBuilderMustValidateNodesExistence(t *testing.T) {
	chainProcessor, err := NewChainProcessorBuilder().
		WithMaxQueueLength(10).
		Build()

	assert.Error(t, err)
	assert.Nil(t, chainProcessor)
}

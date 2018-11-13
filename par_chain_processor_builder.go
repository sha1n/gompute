package gompute

import (
	"container/list"
	"github.com/pkg/errors"
	"runtime"
)

type ChainProcessorBuilder struct {
	maxWorkers     int
	maxQueueLength int
	processNodes   *list.List
}

func NewChainProcessorBuilder() *ChainProcessorBuilder {
	return &ChainProcessorBuilder{
		maxWorkers:   runtime.NumCPU(),
		processNodes: list.New(),
	}
}

func (b *ChainProcessorBuilder) WithMaxWorkers(maxWorkers int) *ChainProcessorBuilder {
	b.maxWorkers = maxWorkers

	return b
}

func (b *ChainProcessorBuilder) WithMaxQueueLength(maxQueueLength int) *ChainProcessorBuilder {
	b.maxQueueLength = maxQueueLength
	return b
}

func (b *ChainProcessorBuilder) AppendProcessNodes(nodes ...ProcessNode) *ChainProcessorBuilder {
	for i := range nodes {
		b.processNodes.PushBack(nodes[i])
	}

	return b
}

func (b *ChainProcessorBuilder) Build() (processor Processor, validationError error) {
	if b.maxWorkers < 1 {
		validationError = errors.New("Max workers must be greater than zero and ideally greater than one.")
	}

	if b.maxQueueLength < 1 {
		validationError = errors.New("Max queue length must be greater than zero and high enough to allow smooth operation.")
	}

	if b.processNodes.Len() < 1 {
		validationError = errors.New("No processing nodes have been set")
	}

	if validationError == nil {
		var i = 0
		nodes := make([]ProcessNode, b.processNodes.Len())
		for e := b.processNodes.Front(); e != nil; e = e.Next() {
			nodes[i] = e.Value.(ProcessNode)
			i += 1
		}

		processor = NewChainProcessor(b.maxWorkers, b.maxQueueLength, nodes)
	}

	return processor, validationError
}

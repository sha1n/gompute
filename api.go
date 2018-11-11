package compute

import (
	"runtime"
	"sync"
)

type Processor interface {
	// Tries to submit an item for processing. Returns false if the processor cannot accept the item.
	Process(item interface{}) (ok bool)
	// Starts accepting items for processing. Before that call any submitted item will be rejected.
	Start()
	// This function shuts down this processor. It will allow already submitted items to finish processing before returning.
	Shutdown()
}

type ProcessNode = func(interface{}) interface{}

func NewChainProcessor(maxQueueLength int, nodes ...ProcessNode) Processor {
	p := parChainProcessor{
		started:               false,
		mx:                    new(sync.RWMutex),
		processNodes:          nodes,
		processNodesBufferMap: make(map[int]chan WorkItem),
		dispatchChannel:       make(chan WorkItem, maxQueueLength),
		shutdownChannel:       make(chan struct{}, runtime.NumCPU()),
		workerWaitGroup:       new(sync.WaitGroup),
	}

	for i := range nodes {
		p.processNodesBufferMap[i] = make(chan WorkItem, maxQueueLength)
	}

	return &p
}

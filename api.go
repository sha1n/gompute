package compute

import (
	"runtime"
	"sync"
)

type Processor interface {
	Process(item interface{}) (ok bool)
	Start()
	Shutdown()
}

type ProcessNode = func(interface{}) interface{}

func NewChainProcessor(maxQueueLength int, nodes ...ProcessNode) Processor {
	p := parChainProcessor{
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

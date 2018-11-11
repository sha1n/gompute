package compute

import (
	"runtime"
	"sync"
	"time"
)

type WorkItem interface {
	Payload() interface{}
	NodeId() int
}

type workItem struct {
	payload interface{}
	nodeId  int
}

func (wi workItem) Payload() interface{} {
	return wi.payload
}

func (wi workItem) NodeId() int {
	return wi.nodeId
}

type parChainProcessor struct {
	processNodes          []ProcessNode
	processNodesBufferMap map[int]chan WorkItem
	dispatchChannel       chan WorkItem
	shutdownChannel       chan struct{}
	workerWaitGroup       *sync.WaitGroup
}

func (p *parChainProcessor) Process(item interface{}) (ok bool) {
	return p.dispatch(&workItem{
		payload: item,
		nodeId:  0,
	})
}

func (p *parChainProcessor) Start() {
	// Start dispatcher routine
	go p.dispatcher()()

	// Start worker routines
	workers := runtime.NumCPU() - 1
	for workers > 0 {
		go p.worker()()
		p.workerWaitGroup.Add(1)
		workers -= 1
	}
}

func (p *parChainProcessor) Shutdown() {
	p.shutdownChannel <- struct{}{}
	p.workerWaitGroup.Wait()

	close(p.dispatchChannel)
	close(p.shutdownChannel)
	p.closeWorkerChannels()
}

func (p *parChainProcessor) dispatch(workedItem WorkItem) (ok bool) {
	return safeEnqueue(p.dispatchChannel, workedItem)
}

func (p *parChainProcessor) busiestQueue() chan WorkItem {
	queues := p.processNodesBufferMap
	maxLen := 0
	var longestQueue = queues[0]
	for k := range queues {
		if len(queues[k]) > maxLen {
			longestQueue = queues[k]
			maxLen = len(longestQueue)
		}
	}

	return longestQueue
}

func (p *parChainProcessor) remainingItems() int {
	queues := p.processNodesBufferMap
	itemCount := 0
	for k := range queues {
		itemCount += len(queues[k])
	}

	return itemCount
}

func (p *parChainProcessor) dispatcher() func() {
	return func() {
		for wi := range p.dispatchChannel {
			safeEnqueue(p.processNodesBufferMap[wi.NodeId()], wi)
		}
	}
}

func (p *parChainProcessor) worker() func() {
	return func() {
		timer := time.NewTicker(time.Millisecond * 10)
		defer p.workerWaitGroup.Done()

		for {
			targetQueue := p.busiestQueue()
			select {
			case wi := <-targetQueue:
				if wi != nil {
					nextHandlerId := wi.NodeId() + 1
					result := p.processNodes[wi.NodeId()](wi.Payload())
					if nextHandlerId < len(p.processNodes) {
						safeEnqueue(p.dispatchChannel, &workItem{
							payload: result,
							nodeId:  nextHandlerId,
						})
					}
				}
			case <-timer.C:
				select { // check for shutdown signal
				case sig := <-p.shutdownChannel:
					p.shutdownChannel <- sig // return to channel for next worker
					if p.remainingItems() == 0 {
						return
					}

				case <-timer.C:
				}
			}
		}

	}
}

func (p *parChainProcessor) closeWorkerChannels() {
	for k := range p.processNodesBufferMap {
		close(p.processNodesBufferMap[k])
	}
}

func safeEnqueue(queue chan WorkItem, wi WorkItem) (ok bool) {
	ok = true
	var rejected = false
	defer func() {
		ok = recover() == nil && !rejected
	}()

	if ok {
		if len(queue) < cap(queue) {
			queue <- wi
		} else {
			rejected = true
		}
	}

	return ok && !rejected

}

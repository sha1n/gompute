package gompute

import (
	"log"
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
	started               bool
	stateMutex            *sync.RWMutex
	maxWorkers            int
	processNodes          []ProcessNode
	processNodesBufferMap map[int]chan WorkItem
	dispatchChannel       chan WorkItem
	shutdownChannel       chan struct{}
	workerWaitGroup       *sync.WaitGroup
}

func NewChainProcessor(maxWorkers int, maxQueueLength int, nodes []ProcessNode) Processor {
	if maxWorkers < 1 {
		log.Panic("Max workers must be greater than zero and ideally greater than one.")
	}
	if maxQueueLength < 1 {
		log.Panic("Max queue length must be greater than zero and high enough to allow smooth operation.")
	}
	if len(nodes) == 0 {
		log.Panic("No processing nodes have been set.")
	}

	p := parChainProcessor{
		started:               false,
		stateMutex:            new(sync.RWMutex),
		maxWorkers:            maxWorkers,
		processNodes:          nodes,
		processNodesBufferMap: make(map[int]chan WorkItem),
		dispatchChannel:       make(chan WorkItem, maxQueueLength),
		shutdownChannel:       make(chan struct{}, maxWorkers),
		workerWaitGroup:       new(sync.WaitGroup),
	}

	for i := range nodes {
		p.processNodesBufferMap[i] = make(chan WorkItem, maxQueueLength)
	}

	return &p
}

func (p *parChainProcessor) Process(item interface{}) (ok bool) {
	if !p.isStarted() {
		ok = false
	} else {
		ok = p.dispatch(&workItem{
			payload: item,
			nodeId:  0,
		})
	}

	return ok
}

func (p *parChainProcessor) Start() {
	if p.isStarted() {
		return
	}

	// Start dispatcher routine
	go p.dispatcher()()

	// Start worker routines
	workers := p.maxWorkers
	for workers > 0 {
		go p.worker()()
		p.workerWaitGroup.Add(1)
		workers -= 1
	}

	p.setStarted(true)
}

func (p *parChainProcessor) Shutdown() {
	if !p.isStarted() {
		return
	}

	p.shutdownChannel <- struct{}{}
	p.workerWaitGroup.Wait()

	close(p.dispatchChannel)
	close(p.shutdownChannel)
	p.closeWorkerChannels()

	p.setStarted(false)
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

func (p *parChainProcessor) isStarted() bool {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()

	return p.started
}

func (p *parChainProcessor) setStarted(state bool) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.started = state
}

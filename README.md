[![Build Status](https://travis-ci.org/sha1n/compute.svg?branch=master)](https://travis-ci.org/sha1n/compute)

# Compute
This repo provides a generic utility for high throughput parallel chain processing. 
Designed for cases where you need to pass many pieces of data through a chain of processing algorithms, while utilizing 
all available CPU cores as much as possible. This implementation uses N-1 go routines to constantly execute processing 
algorithms against the payload and 1 routine to dispatch items submitted by the calling environment, or those returned 
by workers and need to be passed to the next processing algorithm. Below is a high level diagram that tries to illustrate 
the design roughly.      

## High Level Design Diagram
![](docs/par_chain_proc.png?raw=true "Diagram")

## Usage Example
```go
processingNode1 := func(in interface{}) interface{} {
	...
	return output
}

...

processingNodeN := func(in interface{}) interface{} {
	...
	someOutputChannel<- output // it's up to the caller to dispatch the result of the last processing node. 
	return output
}


// bufferSize determines how many items can be enqueued before new items are rejected
chainProcessor := NewChainProcessor(bufferSize, 
    processingNode1,
    ...
    processingNodeN)

// the processor must be started before items can be submitted
chainProcessor.Start()

// the shutdown method will allow already submitted items to go through the entire chain.  
defer chainProcessot.Shutdown()

// Start parallel processing
ok := chainProcessor.Process(input1) // you probably want to check the returned value, to make sure it's not rejected!
...
ok = chainProcessor.Process(input2)
...
ok = chainProcessor.Process(inputN)

```

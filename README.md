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
	someOutputChannel<- output 
	return output
}


chainProcessor := NewChainProcessor(bufferSize,
    processingNode1,
    ...
    processingNodeN)

chainProcessor.Start()
defer chainProcessot.Shutdown()

// Start parallel processing
chainProcessor.Process(input1)
chainProcessor.Process(input2)
...
chainProcessor.Process(inputN)

```

[![Build Status](https://travis-ci.org/sha1n/compute.svg?branch=master)](https://travis-ci.org/sha1n/compute)

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

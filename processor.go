package gompute

type Processor interface {
	// Tries to submit an item for processing. Returns false if the processor cannot accept the item.
	Process(item interface{}) (ok bool)
	// Starts accepting items for processing. Before that call any submitted item will be rejected.
	Start()
	// This function shuts down this processor. It will allow already submitted items to finish processing before returning.
	Shutdown()
}

type ProcessNode = func(interface{}) interface{}

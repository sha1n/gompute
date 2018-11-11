package compute

type Processor interface {
	Process(item interface{}) (ok bool)
	Start()
	Shutdown()
}

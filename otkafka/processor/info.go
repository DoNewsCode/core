package processor

// Info the info of BatchHandler.
//
// Note:
//		If sequence is necessary, make sure that per worker count is one.
//		Multiple goroutines cannot guarantee the order in which data is processed.
type Info struct {
	// used to get reader from otkafka.ReaderMaker.
	Name string
	// reader workers count.
	ReadWorker int
	// batch workers count.
	BatchWorker int
	// data size for batch processing.
	BatchSize int
	// handler workers count.
	HandleWorker int
	// the size of the data channel.
	ChanSize int

	// auto commit *kafka.Message
	// 	true: ignore Handler.Handle or BatchHandler.Batch error info, always commit.
	// 	false: only when Handler.Handle or BatchHandler.Batch error is nil, then commit.
	AutoCommit bool
}

func (i Info) name() string {
	if i.Name == "" {
		return "default"
	}
	return i.Name
}

func (i *Info) readWorker() int {
	if i.ReadWorker <= 0 {
		return 1
	}
	return i.ReadWorker
}

func (i *Info) batchWorker() int {
	if i.BatchWorker <= 0 {
		return 1
	}
	return i.BatchWorker
}

func (i *Info) batchSize() int {
	if i.BatchSize <= 0 {
		return 1
	}
	return i.BatchSize
}

func (i *Info) handleWorker() int {
	if i.HandleWorker <= 0 {
		return 1
	}
	return i.HandleWorker
}

func (i *Info) chanSize() int {
	if i.ChanSize <= 0 {
		return 100
	}
	return i.ChanSize
}

func (i *Info) autoCommit() bool {
	return i.AutoCommit
}

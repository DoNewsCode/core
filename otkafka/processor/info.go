package processor

import "time"

// Info the info of BatchHandler.
//
// Note:
//		If sequence is necessary, make sure that per worker count is one.
//		Multiple goroutines cannot guarantee the order in which data is processed.
type Info struct {
	// used to get reader from otkafka.ReaderMaker.
	// default: "default"
	Name string
	// reader workers count.
	// default: 1
	ReadWorker int
	// batch workers count.
	// default: 1
	BatchWorker int
	// data size for batch processing.
	// default: 1
	BatchSize int
	// handler workers count.
	HandleWorker int
	// the size of the data channel.
	// default: 100
	ChanSize int
	// run the batchFunc automatically at specified intervals, avoid not executing without reaching BatchSize
	// default: 30s
	AutoBatchInterval time.Duration
}

func (i *Info) name() string {
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

func (i *Info) autoBatchInterval() time.Duration {
	if i.AutoBatchInterval < 10 {
		return 30 * time.Second
	}
	return i.AutoBatchInterval
}

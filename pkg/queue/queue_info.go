package queue

// QueueInfo describes the state of queues.
type QueueInfo struct {
	// Waiting is the length of the Waiting queue.
	Waiting int64
	// Delayed is the length of the Delayed queue.
	Delayed int64
	//Timeout is the length of the Timeout queue.
	Timeout int64
	// Failed is the length of the Failed queue.
	Failed int64
}

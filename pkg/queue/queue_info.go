package queue

type QueueInfo struct {
	Waiting int64
	Delayed int64
	Timeout int64
	Failed  int64
}

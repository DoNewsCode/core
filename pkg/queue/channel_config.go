package queue

// ChannelConfig describes the key name of each queue, also known as channel.
type ChannelConfig struct {
	Delayed  string
	Failed   string
	Reserved string
	Waiting  string
	Timeout  string
}

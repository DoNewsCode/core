package leader

// Option is the available options to configure package leader.
type Option struct {
	// The name of the etcd instance.
	EtcdName string `json:"etcdName" yaml:"etcdName"`
}

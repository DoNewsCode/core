package otetcd

import (
	"github.com/DoNewsCode/core/di"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Maker is models Factory
type Maker interface {
	Make(name string) (*clientv3.Client, error)
}

// Factory is a *di.Factory that creates *clientv3.Client using a
// specific configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates *clientv3.Client using a specific configuration entry.
func (r Factory) Make(name string) (*clientv3.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*clientv3.Client), nil
}

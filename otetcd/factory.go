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
type Factory = di.Factory[*clientv3.Client]

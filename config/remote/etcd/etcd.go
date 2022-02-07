// Package etcd allows the core package to bootstrap its configuration from an etcd server.
package etcd

import (
	"context"
	"errors"
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"

	"go.etcd.io/etcd/client/v3"
)

// ETCD is a core.ConfProvider and contract.ConfigWatcher implementation to read and watch remote config key.
// The remote client uses etcd.
type ETCD struct {
	key          string
	clientConfig clientv3.Config
	rev          int64
}

// Provider create a *ETCD
func Provider(clientConfig clientv3.Config, key string) *ETCD {
	return &ETCD{
		key:          key,
		clientConfig: clientConfig,
	}
}

// WithKey is a two-in-one coreOption. It uses the remote key on etcd as the
// source of configuration, and watches the change of that key for hot reloading.
func WithKey(cfg clientv3.Config, key string, codec contract.Codec) (core.CoreOption, core.CoreOption) {
	r := Provider(cfg, key)
	return core.WithConfigStack(r, config.CodecParser{Codec: codec}), core.WithConfigWatcher(r)
}

// ReadBytes reads the contents of a key from etcd and returns the bytes.
func (r *ETCD) ReadBytes() ([]byte, error) {
	client, err := clientv3.New(r.clientConfig)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	resp, err := client.Get(r.context(), r.key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, fmt.Errorf("no such config key: %s", r.key)
	}
	r.rev = resp.Header.Revision
	return resp.Kvs[0].Value, nil
}

// Read is not supported by the remote provider.
func (r *ETCD) Read() (map[string]any, error) {
	return nil, errors.New("remote provider does not support this method")
}

// Watch watches the change to the remote key from etcd. If the key is edited or created, the reload function
// will be called. note the reload function should not just load the changes made within this key, but rather
// it should reload the whole config stack. For example, if the flag or env takes precedence over the config
// key, they should remain to be so after the key changes.
func (r *ETCD) Watch(ctx context.Context, reload func() error) error {
	client, err := clientv3.New(r.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	rch := client.Watch(ctx, r.key, clientv3.WithRev(r.rev))
	for {
		select {
		case resp := <-rch:
			if resp.Err() != nil {
				return resp.Err()
			}
			// Trigger event.
			if err := reload(); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *ETCD) context() context.Context {
	if r.clientConfig.Context != nil {
		return r.clientConfig.Context
	}
	return context.Background()
}

// Package etcd allows the core package to bootstrap its configuration from an etcd server.
package etcd

import (
	"context"
	"errors"
	"fmt"

	"go.etcd.io/etcd/client/v3"
)

// ETCD is a core.ConfProvider and contract.ConfigWatcher implementation to read and watch remote config key.
// The remote client uses etcd.
type ETCD struct {
	key          string
	clientConfig *clientv3.Config
}

// Provider create a *ETCD
func Provider(key string, clientConfig *clientv3.Config) *ETCD {
	return &ETCD{
		key:          key,
		clientConfig: clientConfig,
	}
}

// ReadBytes reads the contents of a key from etcd and returns the bytes.
func (r *ETCD) ReadBytes() ([]byte, error) {
	client, err := clientv3.New(*r.clientConfig)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	resp, err := client.Get(context.Background(), r.key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, fmt.Errorf("no such config key: %s", r.key)
	}

	return resp.Kvs[0].Value, nil
}

// Read is not supported by the remote provider.
func (r *ETCD) Read() (map[string]interface{}, error) {
	return nil, errors.New("remote provider does not support this method")
}

// Watch watches the change to the remote key from etcd. If the key is edited or created, the reload function
// will be called. note the reload function should not just load the changes made within this key, but rather
// it should reload the whole config stack. For example, if the flag or env takes precedence over the config
// key, they should remain to be so after the key changes.
func (r *ETCD) Watch(ctx context.Context, reload func() error) error {
	client, err := clientv3.New(*r.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	rch := client.Watch(ctx, r.key)
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

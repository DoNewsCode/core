package remote

import (
	"context"
	"errors"
	"fmt"

	"go.etcd.io/etcd/client/v3"
)

// Remote is a core.ConfProvider and contract.ConfigWatcher implementation to read and watch remote config file.
// The remote client uses etcd.
type Remote struct {
	path         string
	clientConfig *clientv3.Config
}

// Provider create a *Remote
func Provider(path string, clientConfig *clientv3.Config) *Remote {
	return &Remote{
		path:         path,
		clientConfig: clientConfig,
	}
}

// ReadBytes reads the contents of a file from etcd and returns the bytes.
func (r *Remote) ReadBytes() ([]byte, error) {
	client, err := clientv3.New(*r.clientConfig)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	resp, err := client.Get(context.Background(), r.path)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, fmt.Errorf("no such config path: %s", r.path)
	}

	return resp.Kvs[0].Value, nil
}

// Read is not supported by the remote provider.
func (r *Remote) Read() (map[string]interface{}, error) {
	return nil, errors.New("file provider does not support this method")
}

// Watch watches the change to the remote file from etcd. If the file is edited or created, the reload function
// will be called. note the reload function should not just load the changes made within this file, but rather
// it should reload the whole config stack. For example, if the flag or env takes precedence over the config
// file, they should remain to be so after the file changes.
func (r *Remote) Watch(ctx context.Context, reload func() error) error {
	client, err := clientv3.New(*r.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	rch := client.Watch(ctx, r.path)
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

// Put add values to remote config file on etcd.
func (r *Remote) Put(val string) error {
	client, err := clientv3.New(*r.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Put(context.Background(), r.path, val)
	if err != nil {
		return err
	}
	return nil
}

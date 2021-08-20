package remote

import (
	"context"
	"errors"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	crypt "github.com/DoNewsCode/crypt/config"
	"time"
)

type remote struct {
	cm        crypt.Manager
	key       string
	watchQuit chan bool
}

type Config struct {
	// Name support etcd, firestore and consul.
	Name      string
	Endpoints []string
	Key       string
}

// WithKey is a two-in-one coreOption. It uses the remote key as the
// source of configuration, and watches the change of that key for hot reloading.
func WithKey(cfg Config, codec contract.Codec) (core.CoreOption, core.CoreOption) {
	r := Provider(cfg)
	return core.WithConfigStack(r, config.CodecParser{Codec: codec}), core.WithConfigWatcher(r)
}

// Provider create a core.ConfProvider
func Provider(cfg Config) *remote {
	cm, err := crypt.NewConfigManager(crypt.Config{
		Name:          cfg.Name,
		Machines:      cfg.Endpoints,
		WatchInterval: 10*time.Second,
	})
	if err != nil {
		panic(err)
	}
	quit := make(chan bool)
	return &remote{
		cm:        cm,
		key:       cfg.Key,
		watchQuit: quit,
	}
}

func (r *remote) set(key string, value []byte) error {
	return r.cm.Set(context.TODO(), key, value)
}

// ReadBytes reads the contents of a key and returns the bytes.
func (r *remote) ReadBytes() ([]byte, error) {
	return r.cm.Get(context.TODO(), r.key)
}

// Read is not supported by the remote provider.
func (r *remote) Read() (map[string]interface{}, error) {
	return nil, errors.New("remote provider does not support this method")
}

// Watch watches the change to the remote key. If the key is edited or created, the reload function
// will be called. note the reload function should not just load the changes made within this key, but rather
// it should reload the whole config stack. For example, if the flag or env takes precedence over the config
// key, they should remain to be so after the key changes.
func (r *remote) Watch(ctx context.Context, reload func() error) error {
	rch := r.cm.Watch(ctx,r.key)
	for {
		select {
		case resp := <-rch:
			if resp.Error != nil {
				return resp.Error
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
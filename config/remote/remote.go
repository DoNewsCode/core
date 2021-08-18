package remote

import (
	"context"
	"errors"
	"os"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	crypt "github.com/DoNewsCode/crypt/config"
)

type remote struct {
	cm        crypt.ConfigManager
	key       string
	watchQuit chan bool
}

type Config struct {
	// Name support etcd, firestore and consul.
	Name      string
	Endpoints []string
	Key       string
	SecretKey string
}

// WithKey is a two-in-one coreOption. It uses the remote key as the
// source of configuration, and watches the change of that key for hot reloading.
func WithKey(cfg Config, codec contract.Codec) (core.CoreOption, core.CoreOption) {
	r := Provider(cfg)
	return core.WithConfigStack(r, config.CodecParser{Codec: codec}), core.WithConfigWatcher(r)
}

// Provider create a core.ConfProvider
func Provider(cfg Config) *remote {
	cm, err := getConfigManager(cfg.Name, cfg.Endpoints, cfg.SecretKey)
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
	return r.cm.Set(key, value)
}

// ReadBytes reads the contents of a key and returns the bytes.
func (r *remote) ReadBytes() ([]byte, error) {
	return r.cm.Get(r.key)
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
	defer func() {
		r.watchQuit <- true
	}()
	rch := r.cm.Watch(r.key, r.watchQuit)

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

func getConfigManager(name string, endpoint []string, secretKey string) (crypt.ConfigManager, error) {
	var cm crypt.ConfigManager
	var err error

	if secretKey != "" {
		var kr *os.File
		kr, err = os.Open(secretKey)
		if err != nil {
			return nil, err
		}
		defer kr.Close()
		switch name {
		case "etcd":
			cm, err = crypt.NewEtcdConfigManager(endpoint, kr)
		case "firestore":
			cm, err = crypt.NewFirestoreConfigManager(endpoint, kr)
		default:
			cm, err = crypt.NewConsulConfigManager(endpoint, kr)
		}
	} else {
		switch name {
		case "etcd":
			cm, err = crypt.NewStandardEtcdConfigManager(endpoint)
		case "firestore":
			cm, err = crypt.NewStandardFirestoreConfigManager(endpoint)
		default:
			cm, err = crypt.NewStandardConsulConfigManager(endpoint)
		}
	}
	if err != nil {
		return nil, err
	}
	return cm, nil
}

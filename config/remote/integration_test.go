package remote_test

import (
	"context"
	"github.com/DoNewsCode/core"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_integration(t *testing.T) {
	addr := os.Getenv("ETCD_ADDR")
	if addr == "" {
		t.Skip("set ETCD_ADDR for run remote test")
		return
	}
	key := "core.yaml"
	envEtcdAddrs := strings.Split(addr, ",")
	cfg := clientv3.Config{
		Endpoints:   envEtcdAddrs,
		DialTimeout: time.Second,
	}
	if err := put(cfg, key, "name: remote"); err != nil {
		t.Fatal(err)
	}

	c := core.New(core.WithRemoteYamlFile(key, cfg))
	c.ProvideEssentials()
	assert.Equal(t, "remote", c.String("name"))
}

func put(cfg clientv3.Config, key, val string) error {
	client, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.Put(ctx, key, val)
	if err != nil {
		return err
	}

	return nil
}

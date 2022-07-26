package etcd_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/codec/yaml"
	"github.com/DoNewsCode/core/config/remote/etcd"
	"github.com/DoNewsCode/core/contract"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func Example() {
	addr := os.Getenv("ETCD_ADDR")
	if addr == "" {
		fmt.Println("set ETCD_ADDR for run example")
		return
	}
	key := "core.yaml"
	envEtcdAddrs := strings.Split(addr, ",")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cfg := clientv3.Config{
		Endpoints:   envEtcdAddrs,
		DialTimeout: time.Second,
		Context:     ctx,
	}
	_ = put(cfg, key, "name: etcd")

	c := core.New(etcd.WithKey(cfg, key, yaml.Codec{}))
	c.ProvideEssentials()
	c.Invoke(func(accessor contract.ConfigAccessor) {
		fmt.Println(accessor.String("name"))
	})

	// Output:
	// etcd
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

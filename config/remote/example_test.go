package remote_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/codec/yaml"
	"github.com/DoNewsCode/core/config/remote"
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
	_ = put(clientv3.Config{
		Endpoints:   envEtcdAddrs,
		DialTimeout: 5 * time.Second,
	}, key, "name: etcd")

	cfg := remote.Config{
		Endpoints: envEtcdAddrs,
		Name:      "etcd",
		Key:       key,
	}
	c := core.New(remote.WithKey(cfg, yaml.Codec{}))
	c.ProvideEssentials()
	fmt.Println(c.String("name"))

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

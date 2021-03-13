// +build integration

package otetcd_test

import (
	"context"
	"fmt"
	"log"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otetcd"
	"go.etcd.io/etcd/client/v3"
)

func Example() {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(otetcd.Providers())
	c.Invoke(func(cli *clientv3.Client) {
		_, err := cli.Put(context.TODO(), "foo", "bar")
		if err != nil {
			log.Fatal("etcd put failed")
		}
		resp, _ := cli.Get(context.TODO(), "foo")
		for _, ev := range resp.Kvs {
			fmt.Printf("%s : %s\n", ev.Key, ev.Value)
		}
	})
	// Output:
	// foo : bar
}

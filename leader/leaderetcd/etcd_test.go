package leaderetcd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/DoNewsCode/core/internal"
	"github.com/DoNewsCode/core/key"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

var envDefaultEtcdAddrs, envDefaultEtcdAddrsIsSet = internal.GetDefaultAddrsFromEnv("ETCD_ADDR", "127.0.0.1:2379")

func TestMain(m *testing.M) {
	if !envDefaultEtcdAddrsIsSet {
		fmt.Println("Set env ETCD_ADDR to run leaderetcd tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNewEtcdDriver(t *testing.T) {
	client, _ := clientv3.New(clientv3.Config{Endpoints: envDefaultEtcdAddrs})
	e1 := NewEtcdDriver(client, key.New("test"))
	e2 := NewEtcdDriver(client, key.New("test"))

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan *EtcdDriver)

	go func() {
		e1.Campaign(ctx)
		ch <- e1
	}()
	go func() {
		e2.Campaign(ctx)
		ch <- e2
	}()
	e3 := <-ch
	resp, err := e3.election.Leader(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	e3.Resign(ctx)

	e4 := <-ch
	resp, err = e4.election.Leader(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.NotEqual(t, e3, e4)
	e4.Resign(ctx)
	cancel()
}

package leaderetcd

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core/key"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestNewEtcdDriver(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestNewEtcdDriver")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	client, err := clientv3.New(clientv3.Config{Endpoints: addrs, DialTimeout: 2 * time.Second})
	assert.NoError(t, err)
	defer client.Close()

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

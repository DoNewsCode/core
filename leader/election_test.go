package leader

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/internal"
	"github.com/DoNewsCode/core/key"
	leaderetcd2 "github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/atomic"
)

var envDefaultRedisAddrs, envDefaultRedisAddrsIsSet = internal.GetDefaultAddrsFromEnv("REDIS_ADDR", "127.0.0.1:6379")

func TestMain(m *testing.M) {
	if !envDefaultRedisAddrsIsSet {
		fmt.Println("Set env REDIS_ADDR to run leader tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestElection(t *testing.T) {
	var dispatcher = &events.SyncDispatcher{}
	var e1, e2 Election

	client, err := clientv3.New(clientv3.Config{Endpoints: envDefaultRedisAddrs})
	assert.NoError(t, err)
	e1 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Bool{}},
		driver:     leaderetcd2.NewEtcdDriver(client, key.New("test")),
	}
	e2 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Bool{}},
		driver:     leaderetcd2.NewEtcdDriver(client, key.New("test")),
	}
	ctx, cancel := context.WithCancel(context.Background())

	e1.Campaign(ctx)
	assert.Equal(t, e1.status.IsLeader(), true)

	go e2.Campaign(ctx)
	<-time.After(time.Second)

	assert.Equal(t, e1.status.IsLeader(), true)
	assert.Equal(t, e2.status.IsLeader(), false)

	e1.Resign(ctx)
	time.Sleep(time.Second)
	assert.Equal(t, e1.status.IsLeader(), false)
	assert.Equal(t, e2.status.IsLeader(), true)

	cancel()
}

package leader

import (
	"context"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/key"
	leaderetcd2 "github.com/DoNewsCode/core/leader/leaderetcd"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestElection(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestElection")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	dispatcher := &events.Event[*Status]{}
	var e1, e2 Election

	client, err := clientv3.New(clientv3.Config{Endpoints: addrs, DialTimeout: 2 * time.Second})
	assert.NoError(t, err)
	defer client.Close()

	e1 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Value{}},
		driver:     leaderetcd2.NewEtcdDriver(client, key.New("test")),
	}
	e2 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Value{}},
		driver:     leaderetcd2.NewEtcdDriver(client, key.New("test")),
	}
	ctx, cancel := context.WithCancel(context.Background())

	go e1.Campaign(ctx)
	<-time.After(time.Second)
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

// +build integration

package leader

import (
	"context"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/key"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/atomic"
	"testing"
	"time"
)

func TestElection(t *testing.T) {
	var dispatcher = &events.SyncDispatcher{}
	var e1, e2 Election

	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	assert.NoError(t, err)
	e1 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Bool{}},
		driver: &EtcdDriver{
			keyer:  key.New("test"),
			client: client,
		},
	}
	e2 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Bool{}},
		driver: &EtcdDriver{
			keyer:  key.New("test"),
			client: client,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())

	e1.Campaign(ctx)
	assert.Equal(t, e1.status.IsLeader(), true)

	go e2.Campaign(ctx)
	<-time.After(time.Second)

	assert.Equal(t, e1.status.IsLeader(), true)
	assert.Equal(t, e2.status.IsLeader(), false)

	e1.Resign(ctx)
	assert.Equal(t, e1.status.IsLeader(), false)
	assert.Equal(t, e2.status.IsLeader(), true)

	cancel()
}

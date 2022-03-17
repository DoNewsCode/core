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
	"github.com/DoNewsCode/core/leader/leaderetcd"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestElection(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestElection")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	dispatcher := Dispatcher(&events.Event[*Status]{})

	// confirm status changed event is dispatched
	statusChangedEvent := Status{isLeader: &atomic.Value{}}
	onLeaderChange := dispatcher.(StatusChanged)
	onLeaderChange.On(func(ctx context.Context, status *Status) error {
		statusChangedEvent.isLeader.Store(status.IsLeader())
		return nil
	})

	var e1, e2 Election

	client, err := clientv3.New(clientv3.Config{Endpoints: addrs, DialTimeout: 2 * time.Second})
	assert.NoError(t, err)
	defer client.Close()

	e1 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Value{}},
		driver:     leaderetcd.NewEtcdDriver(client, key.New("el_test")),
	}
	e2 = Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Value{}},
		driver:     leaderetcd.NewEtcdDriver(client, key.New("el_test")),
	}
	ctx, cancel := context.WithCancel(context.Background())

	go e1.Campaign(ctx)
	<-time.After(time.Second)
	assert.Equal(t, true, e1.status.IsLeader())
	assert.Equal(t, true, statusChangedEvent.IsLeader())

	go e2.Campaign(ctx)
	<-time.After(time.Second)

	assert.Equal(t, true, e1.status.IsLeader())
	assert.Equal(t, false, e2.status.IsLeader())

	e1.Resign(ctx)
	time.Sleep(time.Second)
	assert.Equal(t, false, e1.status.IsLeader())
	assert.Equal(t, true, e2.status.IsLeader())

	e2.Resign(ctx)
	time.Sleep(time.Second)
	assert.Equal(t, false, e1.status.IsLeader())
	assert.Equal(t, false, e2.status.IsLeader())
	assert.Equal(t, false, statusChangedEvent.IsLeader())

	cancel()
}

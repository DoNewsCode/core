package leaderetcd_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/leader"
	"github.com/DoNewsCode/core/leader/leaderetcd"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestNewEtcdDriver(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestNewEtcdDriver")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	dispatcher := &events.Event[*leader.Status]{}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	client, err := clientv3.New(clientv3.Config{Endpoints: addrs, DialTimeout: 10 * time.Second, Context: ctx})
	assert.NoError(t, err)
	defer client.Close()

	e1d := leaderetcd.NewEtcdDriver(client, key.New("test"))
	e2d := leaderetcd.NewEtcdDriver(client, key.New("test"))

	e1 := leader.NewElection(dispatcher, e1d)
	e2 := leader.NewElection(dispatcher, e2d)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go e1.Campaign(ctx)
	assert.Eventually(t, func() bool {
		return e1.Status().IsLeader()
	}, 3*time.Second, 10*time.Millisecond, "e1 should be leader")

	go e2.Campaign(ctx)

	assert.Never(t, func() bool {
		return e2.Status().IsLeader()
	}, 3*time.Second, 10*time.Millisecond, "e2 should not be leader")

	e1.Resign(ctx)

	assert.Eventually(t, func() bool {
		return e2.Status().IsLeader()
	}, 3*time.Second, 10*time.Millisecond, "e2 should be leader")
	assert.Never(t, func() bool {
		return e1.Status().IsLeader()
	}, 3*time.Second, 10*time.Millisecond, "e1 should not be leader")

	e2.Resign(ctx)
	cancel()
}

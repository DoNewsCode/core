package leader

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/oklog/run"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/atomic"
)

type In struct {
	di.In

	dispatcher contract.Dispatcher
	client     *clientv3.Client
}

type Out struct {
	di.Out
	di.Module

	Election Election
	Status   Status
}

func Provide(in In) Out {
	return Out{
		Election: Election{
			dispatcher: in.dispatcher,
			status:     &Status{isLeader: &atomic.Bool{}},
			driver: EtcdDriver{
				client:  in.client,
				session: nil,
			},
		},
	}
}

func (m Out) ProvideCloser() {
	_ = m.Election.Resign(context.Background())
}

func (m Out) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(func() error {
		return m.Election.Campaign(ctx)
	}, func(err error) {
		cancel()
	})
}

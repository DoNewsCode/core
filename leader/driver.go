package leader

import (
	"context"
	"os"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type EtcdDriver struct {
	client  *clientv3.Client
	session *concurrency.Session
}

func (e EtcdDriver) Campaign(ctx context.Context) error {
	var err error
	e.session, err = concurrency.NewSession(e.client)
	if err != nil {
		return err
	}
	c := concurrency.NewElection(e.session, "/leader-election/")
	hostname, _ := os.Hostname()
	if err := c.Campaign(ctx, hostname); err != nil {
		return err
	}
	return nil
}

func (e EtcdDriver) Resign(ctx context.Context) error {
	defer e.session.Close()
	return e.Resign(ctx)
}

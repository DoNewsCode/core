package leaderetcd

import (
	"context"
	"os"

	"github.com/DoNewsCode/core/contract"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

// EtcdDriver implements the Driver. It is used for leader election.
type EtcdDriver struct {
	keyer    contract.Keyer
	client   *clientv3.Client
	session  *concurrency.Session
	election *concurrency.Election
}

// NewEtcdDriver returns a newly constructed *EtcdDriver
func NewEtcdDriver(client *clientv3.Client, keyer contract.Keyer) *EtcdDriver {
	return &EtcdDriver{
		keyer:    keyer,
		client:   client,
		session:  nil,
		election: nil,
	}
}

// Campaign starts the leader election using ETCD. It will bock until this node becomes leader or the context is expired.
func (e *EtcdDriver) Campaign(ctx context.Context) error {
	var err error
	e.session, err = concurrency.NewSession(e.client)
	if err != nil {
		return err
	}
	e.election = concurrency.NewElection(e.session, "/"+e.keyer.Key("/", "leader-election"))
	hostname, _ := os.Hostname()
	if err := e.election.Campaign(ctx, hostname); err != nil {
		return err
	}
	return nil
}

// Resign gives up the leadership using ETCD. If the current node is not a leader, this is a no op.
func (e *EtcdDriver) Resign(ctx context.Context) error {
	if e.session == nil || e.election == nil {
		return nil
	}
	defer e.session.Close()
	return e.election.Resign(ctx)
}

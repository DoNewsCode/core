package mysqlstore

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/dtx/sagas"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
)

/*
Providers returns MySQLStore dependency.
	DependsOn:
		- otgorm.Maker
		- contract.ConfigAccessor
	Provides:
		- *MySQLStore
		- sagas.Store
*/
func Providers() di.Deps {
	return []interface{}{provide, provideConfig}
}

func provide(in in) (out, error) {
	var conf configuration
	err := in.Conf.Unmarshal("sagas-mysql", &conf)
	if err != nil {
		level.Warn(in.Logger).Log("err", err)
	}

	conn := conf.getConnection()

	db, err := in.Maker.Make(conn)
	if err != nil {
		return out{}, err
	}
	var opts []Option
	retention := conf.getRetention().Duration
	cleanupInterval := conf.getCleanupInterval().Duration

	opts = append(opts, WithRetention(retention), WithCleanUpInterval(cleanupInterval))

	store := New(db, opts...)
	return out{
		Conn:      conn,
		Store:     store,
		SagaStore: store,
	}, nil
}

type in struct {
	di.In

	Logger log.Logger
	Maker  otgorm.Maker
	Conf   contract.ConfigAccessor
}

type out struct {
	di.Out

	Conn      string `name:"mysqlstore"`
	Store     *MySQLStore
	SagaStore sagas.Store
}

func (m out) ModuleSentinel() {}

func (m out) ProvideMigration() []*otgorm.Migration {
	return Migrations(m.Conn)
}

func (m out) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(func() error {
		return m.Store.CleanUp(ctx)
	}, func(err error) {
		cancel()
	})
}

package mysqlstore

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/dtx/sagas"
	"github.com/DoNewsCode/core/otgorm"
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
	return []interface{}{provide}
}

func provide(in in) (out, error) {
	conn := "default"
	if in.Conf.String("sagas.mysql.connection") != "" {
		conn = in.Conf.String("sagas.mysql.connection")
	}
	db, err := in.Maker.Make(conn)
	if err != nil {
		return out{}, err
	}
	var opts []Option
	d, _ := time.ParseDuration(in.Conf.String("sagas.mysql.retention"))
	if d > 0 {
		opts = append(
			opts,
			WithRetention(d),
		)
	}
	d, _ = time.ParseDuration(in.Conf.String("sagas.mysql.cleanupInterval"))
	if d > 0 {
		opts = append(
			opts,
			WithCleanUpInterval(d),
		)
	}
	store := New(db, opts...)
	return out{
		Conn:      conn,
		Store:     store,
		SagaStore: store,
	}, nil
}

type in struct {
	di.In

	Maker otgorm.Maker
	Conf  contract.ConfigAccessor
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

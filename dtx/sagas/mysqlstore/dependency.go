package mysqlstore

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/dtx/sagas"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/robfig/cron/v3"
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
	if in.Conf.Float64("sagas.mysql.retentionHour") > 0 {
		opts = append(opts, WithRetention(time.Hour*time.Duration(in.Conf.Float64("sagas.mysql.retentionHour"))))
	}
	store := New(db, opts...)
	return out{
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

	Store     *MySQLStore
	SagaStore sagas.Store
}

func (m out) ModuleSentinel() {}

func (m out) ProvideMigration() []*otgorm.Migration {
	return Migrations()
}

func (m out) ProvideCron(crontab *cron.Cron) {
	crontab.AddFunc("0 2 * * *", func() {
		m.Store.CleanUp(context.Background())
	})
}

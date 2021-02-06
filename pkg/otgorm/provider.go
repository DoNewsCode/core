package otgorm

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type databaseConf struct {
	Database                                 string
	Dsn                                      string
	SkipDefaultTransaction                   bool
	FullSaveAssociations                     bool
	DryRun                                   bool
	PrepareStmt                              bool
	DisableAutomaticPing                     bool
	DisableForeignKeyConstraintWhenMigrating bool
	DisableNestedTransaction                 bool
	AllowGlobalUpdate                        bool
	QueryFields                              bool
	CreateBatchSize                          int
	NamingStrategy                           struct {
		TablePrefix   string
		SingularTable bool
	}
}

func ProvideDialector(conf *databaseConf) (gorm.Dialector, error) {
	if conf.Database == "mysql" {
		return mysql.Open(conf.Dsn), nil
	}
	if conf.Database == "sqlite" {
		return sqlite.Open(conf.Dsn), nil
	}
	return nil, fmt.Errorf("unknow database type %s", conf.Database)
}

func ProvideGormConfig(l log.Logger, conf *databaseConf) *gorm.Config {
	return &gorm.Config{
		SkipDefaultTransaction: conf.SkipDefaultTransaction,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.NamingStrategy.TablePrefix,
			SingularTable: conf.NamingStrategy.SingularTable,
		},
		FullSaveAssociations:                     conf.FullSaveAssociations,
		Logger:                                   &GormLogAdapter{Logging: l},
		DryRun:                                   conf.DryRun,
		PrepareStmt:                              conf.PrepareStmt,
		DisableAutomaticPing:                     conf.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: conf.DisableForeignKeyConstraintWhenMigrating,
		DisableNestedTransaction:                 conf.DisableNestedTransaction,
		AllowGlobalUpdate:                        conf.AllowGlobalUpdate,
		QueryFields:                              conf.QueryFields,
		CreateBatchSize:                          conf.CreateBatchSize,
	}
}

func ProvideGormDB(dialector gorm.Dialector, config *gorm.Config, tracer opentracing.Tracer) (*gorm.DB, func(), error) {
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, nil, err
	}
	if tracer != nil {
		AddGormCallbacks(db, tracer)
	}
	return db, func() {
		if sqlDb, err := db.DB(); err == nil {
			sqlDb.Close()
		}
	}, nil
}

type DatabaseParams struct {
	dig.In

	Conf   contract.ConfigAccessor
	Logger log.Logger
	Tracer opentracing.Tracer `optional:"true"`
}

func ProvideDefaultDatabase(p DatabaseParams) (*gorm.DB, func(), error) {
	factory, _ := ProvideDBFactory(p)
	conn, err := factory.Make("default")
	return conn, func() {
		factory.CloseConn("default")
	}, err
}

type DBFactory struct {
	*async.Factory
}

func (d DBFactory) Make(name string) (*gorm.DB, error) {
	db, err := d.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return db.(*gorm.DB), nil
}

// ProvideMemoryDatabase provides a sqlite database in memory mode. This is useful for testing.
func ProvideMemoryDatabase() *gorm.DB {
	factory, _ := ProvideDBFactory(DatabaseParams{
		In: dig.In{},
		Conf: config.MapAdapter{"gorm": map[string]databaseConf{
			"memory": {
				Database: "sqlite",
				Dsn:      "file::memory:?cache=shared",
			},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	memoryDatabase, _ := factory.Make("memory")
	return memoryDatabase
}

func ProvideDBFactory(p DatabaseParams) (DBFactory, func()) {
	logger := log.With(p.Logger, "component", "database")

	var dbConfs map[string]databaseConf
	err := p.Conf.Unmarshal("gorm", &dbConfs)
	if err != nil {
		level.Warn(logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			dialector gorm.Dialector
			conf      databaseConf
			ok        bool
			conn      *gorm.DB
			cleanup   func()
		)
		if conf, ok = dbConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("database configuration %s not found", name)
		}
		dialector, err = ProvideDialector(&conf)
		if err != nil {
			return async.Pair{}, err
		}
		gormConfig := ProvideGormConfig(logger, &conf)
		conn, cleanup, err = ProvideGormDB(dialector, gormConfig, p.Tracer)
		if err != nil {
			return async.Pair{}, err
		}
		return async.Pair{
			Conn:   conn,
			Closer: cleanup,
		}, err
	})
	dbFactory := DBFactory{factory}
	return dbFactory, dbFactory.Close
}

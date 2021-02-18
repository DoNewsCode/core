package otgorm

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type databaseConf struct {
	Database                                 string `json:"database" yaml:"database"`
	Dsn                                      string `json:"dsn" yaml:"dsn"`
	SkipDefaultTransaction                   bool   `json:"skipDefaultTransaction" yaml:"skipDefaultTransaction"`
	FullSaveAssociations                     bool   `json:"fullSaveAssociations" yaml:"fullSaveAssociations"`
	DryRun                                   bool   `json:"dryRun" yaml:"dryRun"`
	PrepareStmt                              bool   `json:"prepareStmt" yaml:"prepareStmt"`
	DisableAutomaticPing                     bool   `json:"disableAutomaticPing" yaml:"disableAutomaticPing"`
	DisableForeignKeyConstraintWhenMigrating bool   `json:"disableForeignKeyConstraintWhenMigrating" yaml:"disableForeignKeyConstraintWhenMigrating"`
	DisableNestedTransaction                 bool   `json:"disableNestedTransaction" yaml:"disableNestedTransaction"`
	AllowGlobalUpdate                        bool   `json:"allowGlobalUpdate" yaml:"allowGlobalUpdate"`
	QueryFields                              bool   `json:"queryFields" yaml:"queryFields"`
	CreateBatchSize                          int    `json:"createBatchSize" yaml:"createBatchSize"`
	NamingStrategy                           struct {
		TablePrefix   string `json:"tablePrefix" yaml:"tablePrefix"`
		SingularTable bool   `json:"singularTable" yaml:"singularTable"`
	} `json:"namingStrategy" yaml:"namingStrategy"`
}

// ProvideDialector provides a gorm.Dialector. Mean to be used as an intermediate
// step to create *gorm.DB
func ProvideDialector(conf *databaseConf) (gorm.Dialector, error) {
	if conf.Database == "mysql" {
		return mysql.Open(conf.Dsn), nil
	}
	if conf.Database == "sqlite" {
		return sqlite.Open(conf.Dsn), nil
	}
	return nil, fmt.Errorf("unknow database type %s", conf.Database)
}

// ProvideGormConfig provides a *gorm.Config. Mean to be used as an intermediate
// step to create *gorm.DB
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

// ProvideGormDB provides a *gorm.DB. It is intended to be used with
// ProvideDialector and ProvideGormConfig
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

// GormConfigInterceptor is a function that allows user to make last minute
// change to *gorm.Config when constructing *gorm.DB.
type GormConfigInterceptor func(name string, conf *gorm.Config)

// DatabaseIn is the injection parameter for ProvideDatabase.
type DatabaseIn struct {
	di.In

	Conf                  contract.ConfigAccessor
	Logger                log.Logger
	GormConfigInterceptor GormConfigInterceptor `optional:"true"`
	Tracer                opentracing.Tracer    `optional:"true"`
}

// Maker models Factory
type Maker interface {
	Make(name string) (*gorm.DB, error)
}

// DatabaseOut is the result of ProvideDatabase. *gorm.DB is not a interface
// type. It is up to the users to define their own database repository interface.
type DatabaseOut struct {
	di.Out

	Database *gorm.DB
	Factory  Factory
	Maker    Maker
}

// ProvideConfig exports the default database configuration.
func (d DatabaseOut) ProvideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{
		{
			Name: "database",
			Data: map[string]interface{}{
				"database": map[string]databaseConf{
					"default": {
						Database:                                 "mysql",
						Dsn:                                      "root@tcp(127.0.0.1:3306)/skeleton?charset=utf8mb4&parseTime=True&loc=Local",
						SkipDefaultTransaction:                   false,
						FullSaveAssociations:                     false,
						DryRun:                                   false,
						PrepareStmt:                              false,
						DisableAutomaticPing:                     false,
						DisableForeignKeyConstraintWhenMigrating: false,
						DisableNestedTransaction:                 false,
						AllowGlobalUpdate:                        false,
						QueryFields:                              false,
						CreateBatchSize:                          0,
						NamingStrategy: struct {
							TablePrefix   string `json:"tablePrefix" yaml:"tablePrefix"`
							SingularTable bool   `json:"singularTable" yaml:"singularTable"`
						}{},
					},
				},
			},
			Comment: "The database configuration",
		},
	}
}

// ProvideDatabase creates Factory and *gorm.DB. It is a valid dependency for
// package core.
func ProvideDatabase(p DatabaseIn) (DatabaseOut, func()) {
	factory, cleanup := provideDBFactory(p)
	database, err := factory.Make("default")
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	return DatabaseOut{
		Database: database,
		Factory:  factory,
		Maker:    factory,
	}, cleanup
}

// Factory is the *async.Factory that creates *gorm.DB under a specific
// configuration entry.
type Factory struct {
	*async.Factory
}

// Make creates *gorm.DB under a specific configuration entry.
func (d Factory) Make(name string) (*gorm.DB, error) {
	db, err := d.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return db.(*gorm.DB), nil
}

// ProvideMemoryDatabase provides a sqlite database in memory mode. This is
// useful for testing.
func ProvideMemoryDatabase() *gorm.DB {
	factory, _ := provideDBFactory(DatabaseIn{
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

func provideDBFactory(p DatabaseIn) (Factory, func()) {
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
		if p.GormConfigInterceptor != nil {
			p.GormConfigInterceptor(name, gormConfig)
		}
		conn, cleanup, err = ProvideGormDB(dialector, gormConfig, p.Tracer)
		if err != nil {
			return async.Pair{}, err
		}
		return async.Pair{
			Conn:   conn,
			Closer: cleanup,
		}, err
	})
	dbFactory := Factory{factory}
	return dbFactory, dbFactory.Close
}

package otgorm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/eventsv2"
	"github.com/go-kit/log"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

/*
Providers returns a set of database related providers for package core. It includes
the Maker, database configs and the default *gorm.DB instance.
	Depends On:
		contract.ConfigAccessor
		log.Logger
		opentracing.Tracer    `optional:"true"`
		Gauges `optional:"true"`
		contract.Dispatcher `optional:"true"`
	Provide:
		Maker
		Factory
		*gorm.DB
*/
func Providers(opt ...ProvidersOptionFunc) di.Deps {
	o := providersOption{
		interceptor: func(name string, conf *gorm.Config) {},
		drivers:     getDefaultDrivers(),
	}
	for _, f := range opt {
		f(&o)
	}
	return di.Deps{
		provideConfig,
		provideDefaultDatabase,
		provideDBFactory(&o),
		di.Bind(new(*Factory), new(Maker)),
	}
}

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

type metricsConf struct {
	Interval config.Duration `json:"interval" yaml:"interval"`
}

// factoryIn is the injection parameter for provideDatabaseOut.
type factoryIn struct {
	di.In

	Conf          contract.ConfigUnmarshaler
	Logger        log.Logger
	Tracer        opentracing.Tracer      `optional:"true"`
	Gauges        *Gauges                 `optional:"true"`
	OnReloadEvent *eventsv2.OnReloadEvent `optional:"true"`
}

// databaseOut is the result of provideDatabaseOut. *gorm.DB is not a interface
// type. It is up to the users to define their own database repository interface.
type databaseOut struct {
	di.Out

	Factory   *Factory
	Collector *collector
}

// Module implements di.Modular
func (d databaseOut) Module() interface{} {
	return d
}

// ProvideRunGroup implements RunGroupProvider
func (d databaseOut) ProvideRunGroup(group *run.Group) {
	if d.Collector == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(d.Collector.interval)
	group.Add(func() error {
		for {
			select {
			case <-ticker.C:
				d.Collector.collectConnectionStats()
			case <-ctx.Done():
				ticker.Stop()
				return nil
			}
		}
	}, func(err error) {
		cancel()
	})
}

// provideDialector provides a gorm.Dialector. Mean to be used as an intermediate
// step to create *gorm.DB
func provideDialector(conf *databaseConf, drivers Drivers) (gorm.Dialector, error) {
	if driver, ok := drivers[conf.Database]; ok {
		return driver(conf.Dsn), nil
	}
	return nil, fmt.Errorf("unknow database type %s", conf.Database)
}

// provideGormConfig provides a *gorm.Config. Mean to be used as an intermediate
// step to create *gorm.DB
func provideGormConfig(l log.Logger, conf *databaseConf) *gorm.Config {
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

// provideGormDB provides a *gorm.DB. It is intended to be used with
// provideDialector and provideGormConfig. Gorm opens connection to database
// while building *gorm.db. This means if the database is not available, the system
// will fail when initializing dependencies.
func provideGormDB(dialector gorm.Dialector, config *gorm.Config, tracer opentracing.Tracer) (*gorm.DB, func(), error) {
	db, err := gorm.Open(dialector, config)

	var nerr *net.OpError

	if err != nil && !errors.As(err, &nerr) {
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

func provideDefaultDatabase(maker Maker) (*gorm.DB, error) {
	return maker.Make("default")
}

func provideDBFactory(options *providersOption) func(p factoryIn) (databaseOut, func(), error) {
	return func(factoryIn factoryIn) (databaseOut, func(), error) {
		logger := log.With(factoryIn.Logger, "tag", "database")
		if options.drivers == nil {
			options.drivers = getDefaultDrivers()
		}
		if options.interceptor == nil {
			options.interceptor = func(name string, conf *gorm.Config) {}
		}

		factory := di.NewFactory[*gorm.DB](func(name string) (pair di.Pair[*gorm.DB], err error) {
			var (
				dialector gorm.Dialector
				conf      databaseConf
				conn      *gorm.DB
				cleanup   func()
			)
			factoryIn := factoryIn
			if err := factoryIn.Conf.Unmarshal(fmt.Sprintf("gorm.%s", name), &conf); err != nil {
				return pair, fmt.Errorf("database configuration %s not valid: %w", name, err)
			}
			dialector, err = provideDialector(&conf, options.drivers)
			if err != nil {
				return pair, err
			}
			gormConfig := provideGormConfig(logger, &conf)
			options.interceptor(name, gormConfig)
			conn, cleanup, err = provideGormDB(dialector, gormConfig, factoryIn.Tracer)
			if err != nil {
				return pair, err
			}
			return di.Pair[*gorm.DB]{
				Conn:   conn,
				Closer: cleanup,
			}, err
		})
		if options.reloadable && factoryIn.OnReloadEvent != nil {
			factoryIn.OnReloadEvent.Subscribe(func(_ context.Context, _ eventsv2.OnReloadPayload) error {
				factory.Close()
				return nil
			})
		}

		var collector *collector
		if factoryIn.Gauges != nil {
			var interval time.Duration = 15 * time.Second
			factoryIn.Conf.Unmarshal("gormMetrics.interval", &interval)
			collector = newCollector(factory, factoryIn.Gauges, interval)
		}
		return databaseOut{
			Factory:   factory,
			Collector: collector,
		}, factory.Close, nil
	}
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

// ProvideConfig exports the default database configuration.
func provideConfig() configOut {
	exported := []config.ExportedConfig{
		{
			Owner: "otgorm",
			Data: map[string]interface{}{
				"gorm": map[string]databaseConf{
					"default": {
						Database:                                 "mysql",
						Dsn:                                      "root@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local",
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
				"gormMetrics": metricsConf{
					Interval: config.Duration{Duration: 15 * time.Second},
				},
			},
			Comment: "The database configuration",
		},
	}
	return configOut{Config: exported}
}

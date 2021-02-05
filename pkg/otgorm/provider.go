package otgorm

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/dig"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sync"
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
	AddGormCallbacks(db, tracer)
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

func Database(p DatabaseParams) (*gorm.DB, func(), error) {
	var dbConf databaseConf
	err := p.Conf.Unmarshal("gorm.default", &dbConf)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to parse default config")
	}
	dialector, err := ProvideDialector(&dbConf)
	if err != nil {
		return nil, nil, err
	}
	logger := log.With(p.Logger, "component", "database")
	gormConfig := ProvideGormConfig(logger, &dbConf)
	if p.Tracer == nil {
		p.Tracer = opentracing.NoopTracer{}
	}
	return ProvideGormDB(dialector, gormConfig, p.Tracer)
}

type DatabaseFactory struct {
	db map[string]*gorm.DB
}

func NewDatabaseFactory(p DatabaseParams) (*DatabaseFactory, func(), error) {
	var (
		logger    = log.With(p.Logger, "component", "database")
		cleanups  []func()
		err       error
		dialector gorm.Dialector
		g         *gorm.DB
		c         func()
	)

	if p.Tracer == nil {
		p.Tracer = opentracing.NoopTracer{}
	}

	var dbConfs map[string]databaseConf
	err = p.Conf.Unmarshal("gorm", &dbConfs)
	if err != nil {
		return nil, nil, err
	}
	databaseFactory := &DatabaseFactory{
		db: make(map[string]*gorm.DB),
	}
	for name, value := range dbConfs {
		dialector, err = ProvideDialector(&value)
		if err != nil {
			return nil, nil, err
		}
		gormConfig := ProvideGormConfig(logger, &value)
		g, c, err = ProvideGormDB(dialector, gormConfig, p.Tracer)
		if err != nil {
			return nil, nil, err
		}
		databaseFactory.db[name] = g
		cleanups = append(cleanups, c)
	}
	return databaseFactory, func() {
		var wg sync.WaitGroup
		for i := range cleanups {
			wg.Add(1)
			go func(i int) {
				cleanups[i]()
				wg.Done()
			}(i)
		}
		wg.Wait()
	}, nil
}

func (d *DatabaseFactory) Connection(name string) *gorm.DB {
	return d.db[name]
}

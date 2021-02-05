package otgorm

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DatabaseConf struct {
	Database    string
	Dsn         string
	TablePrefix string
}

func ProvideDialector(conf *DatabaseConf) (gorm.Dialector, error) {
	if conf.Database == "mysql" {
		return mysql.Open(conf.Dsn), nil
	}
	if conf.Database == "sqlite" {
		return sqlite.Open(conf.Dsn), nil
	}
	return nil, fmt.Errorf("unknow database type %s", conf.Database)
}

func ProvideGormConfig(l log.Logger, conf *DatabaseConf) *gorm.Config {
	return &gorm.Config{
		Logger:                                   &GormLogAdapter{Logging: l},
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: conf.TablePrefix,
		},
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
	var dbConf DatabaseConf
	_ = p.Conf.Unmarshal("gorm.default.database", &dbConf.Database)
	_ = p.Conf.Unmarshal("gorm.default.dsn", &dbConf.Dsn)
	_ = p.Conf.Unmarshal("gorm.default.tablePrefix", &dbConf.TablePrefix)
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

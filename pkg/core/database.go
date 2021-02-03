package core

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/otgorm"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type DatabaseParams struct {
	dig.In

	Conf   contract.ConfigAccessor
	Logger log.Logger
	Tracer opentracing.Tracer `optional:"true"`
}

func Database(p DatabaseParams) (*gorm.DB, func(), error) {
	var dbConf otgorm.DatabaseConf
	_ = p.Conf.Unmarshal("gorm.database", &dbConf.Database)
	_ = p.Conf.Unmarshal("gorm.dsn", &dbConf.Dsn)
	_ = p.Conf.Unmarshal("gorm.tablePrefix", &dbConf.TablePrefix)
	dialector, err := otgorm.ProvideDialector(&dbConf)
	if err != nil {
		return nil, nil, err
	}
	logger := log.With(p.Logger, "component", "database")
	gormConfig := otgorm.ProvideGormConfig(logger, &dbConf)
	if p.Tracer == nil {
		p.Tracer = opentracing.NoopTracer{}
	}
	return otgorm.ProvideGormDB(dialector, gormConfig, p.Tracer)
}

func ProvideDatabase(c *C) {
	c.Provide(func(p DatabaseParams) (*gorm.DB, error) {
		var dbConf otgorm.DatabaseConf
		_ = p.Conf.Unmarshal("gorm.database", &dbConf.Database)
		_ = p.Conf.Unmarshal("gorm.dsn", &dbConf.Dsn)
		_ = p.Conf.Unmarshal("gorm.tablePrefix", &dbConf.TablePrefix)
		dialector, err := otgorm.ProvideDialector(&dbConf)
		if err != nil {
			return nil, err
		}
		logger := log.With(p.Logger, "component", "database")
		gormConfig := otgorm.ProvideGormConfig(logger, &dbConf)
		if p.Tracer == nil {
			p.Tracer = opentracing.NoopTracer{}
		}
		db, cleanup, err := otgorm.ProvideGormDB(dialector, gormConfig, p.Tracer)
		if err != nil {
			return nil, err
		}
		c.CloserProviders = append(c.CloserProviders, cleanup)
		return db, nil
	})
}

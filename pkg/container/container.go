package container

import (
	"github.com/robfig/cron/v3"
)

type Container struct {
	BaseContainer
	MigrationProviders []Migrations
	SeedProviders      []func() error
	CronProviders      []func(crontab *cron.Cron)
}

type Migrations struct {
	Migrate  func() error
	Rollback func(flag string) error
}

type MigrationProvider interface {
	ProvideMigration() error
	ProvideRollback(flag string) error
}

type SeedProvider interface {
	ProvideSeed() error
}

type CronProvider interface {
	ProvideCron(crontab *cron.Cron)
}

func (s *Container) Register(module interface{}) {
	s.BaseContainer.Register(module)
	if p, ok := module.(MigrationProvider); ok {
		s.MigrationProviders = append(s.MigrationProviders, Migrations{p.ProvideMigration, p.ProvideRollback})
	}
	if p, ok := module.(SeedProvider); ok {
		s.SeedProviders = append(s.SeedProviders, p.ProvideSeed)
	}
	if p, ok := module.(CronProvider); ok {
		s.CronProviders = append(s.CronProviders, p.ProvideCron)
	}
}

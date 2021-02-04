package container

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

type Container struct {
	BaseContainer
	MigrationProviders []Migrations
	SeedProviders      []func() error
	CronProviders      []func(crontab *cron.Cron)
	CommandProviders   []func(command *cobra.Command)
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

type CommandProvider interface {
	ProvideCommand(command *cobra.Command)
}

func (s *Container) AddModule(module interface{}) {
	s.BaseContainer.AddModule(module)
	if p, ok := module.(MigrationProvider); ok {
		s.MigrationProviders = append(s.MigrationProviders, Migrations{p.ProvideMigration, p.ProvideRollback})
	}
	if p, ok := module.(SeedProvider); ok {
		s.SeedProviders = append(s.SeedProviders, p.ProvideSeed)
	}
	if p, ok := module.(CronProvider); ok {
		s.CronProviders = append(s.CronProviders, p.ProvideCron)
	}
	if p, ok := module.(CommandProvider); ok {
		s.CommandProviders = append(s.CommandProviders, p.ProvideCommand)
	}
}

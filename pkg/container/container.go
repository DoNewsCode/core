package container

import (
	"github.com/DoNewsCode/std/pkg/otgorm"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

type Container struct {
	BaseContainer
	MigrationProviders []func() []*otgorm.Migration
	SeedProviders      []func() []*otgorm.Seed
	CronProviders      []func(crontab *cron.Cron)
	CommandProviders   []func(command *cobra.Command)
}

type MigrationProvider interface {
	ProvideMigration() []*otgorm.Migration
}

type SeedProvider interface {
	ProvideSeed() []*otgorm.Seed
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
		s.MigrationProviders = append(s.MigrationProviders, p.ProvideMigration)
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

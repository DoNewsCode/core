package otredis

import (
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/log"
	"github.com/spf13/cobra"
)

// Module is the registration unit for package core. It provides redis command.
type Module struct {
	maker  Maker
	logger log.Logger
}

// ModuleIn contains the input parameters needed for creating the new module.
type ModuleIn struct {
	di.In

	Maker  Maker
	Logger log.Logger
}

// New creates a Module.
func New(in ModuleIn) Module {
	return Module{
		maker:  in.Maker,
		logger: in.Logger,
	}
}

// ProvideCommand provides migration and seed command.
func (m Module) ProvideCommand(command *cobra.Command) {
	cleanupCmd := NewCleanupCommand(m.maker, m.logger)
	redisCmd := &cobra.Command{
		Use:   "redis",
		Short: "manage redis",
		Long:  "manage redis, such as cleaning up redis cache",
	}
	redisCmd.AddCommand(cleanupCmd)
	command.AddCommand(redisCmd)
}

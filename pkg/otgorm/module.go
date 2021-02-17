package otgorm

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
)

// MigrationProvider is an interface for database migrations. Modules
// implementing this interface are migration providers. migrations will be
// collected in migrate command.
type MigrationProvider interface {
	ProvideMigration() []*Migration
}

// SeedProvider is an interface for database seeding. Modules
// implementing this interface are seed providers. seeds will be
// collected in seed command.
type SeedProvider interface {
	ProvideSeed() []*Seed
}

// Module is the registration unit for Morph. It provides migration and seed command.
type Module struct {
	maker     Maker
	env       contract.Env
	logger    log.Logger
	container contract.Container
}

// New creates Module
func New(make Maker, env contract.Env, logger log.Logger, container contract.Container) Module {
	return Module{
		maker:     make,
		env:       env,
		logger:    logger,
		container: container,
	}
}

// ProvideCommand provides migration and seed command.
func (m Module) ProvideCommand(command *cobra.Command) {
	var (
		force      bool
		rollbackId string
		logger     = logging.WithLevel(m.logger)
	)
	var migrateCmd = &cobra.Command{
		Use:   "migrate [database]",
		Short: "Migrate gorm tables",
		Long:  `Run all gorm table migrations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var connection = "default"
			if len(args) > 0 {
				connection = args[0]
			}

			if m.env.IsProduction() && !force {
				e := fmt.Errorf("migrations and rollback in production requires force flag to be set")
				return e
			}

			migrations := m.collectMigrations(connection)

			if rollbackId != "" {
				if err := migrations.Rollback(rollbackId); err != nil {
					return fmt.Errorf("unable to rollback: %w", err)
				}

				logger.Info("rollback successfully completed")
				return nil
			}

			if err := migrations.Migrate(); err != nil {
				return fmt.Errorf("unable to migrate: %w", err)
			}

			logger.Info("migration successfully completed")
			return nil
		},
	}
	migrateCmd.Flags().BoolVarP(&force, "force", "f", false, "migrations and rollback in production requires force flag to be set")
	migrateCmd.Flags().StringVarP(&rollbackId, "rollback", "r", "", "rollback to the given migration id")
	migrateCmd.Flag("rollback").NoOptDefVal = "-1"
	command.AddCommand(migrateCmd)

	var seedCmd = &cobra.Command{
		Use:   "seed [database]",
		Short: "seed the database",
		Long:  `use the provided seeds to bootstrap fake data in database`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var connection = "default"
			if len(args) > 0 {
				connection = args[0]
			}

			if m.env.IsProduction() && !force {
				return fmt.Errorf("seeding in production requires force flag to be set")
			}

			seeds := m.collectSeeds(connection)

			if err := seeds.Seed(); err != nil {
				return fmt.Errorf("seed failed: %w", err)
			}

			logger.Info("seeding successfully completed")
			return nil
		},
	}
	seedCmd.Flags().BoolVarP(&force, "force", "f", false, "seeding in production requires force flag to be set")
	command.AddCommand(seedCmd)
}

func (m Module) collectMigrations(connection string) Migrations {
	if connection == "" {
		connection = "default"
	}
	var migrations Migrations
	m.container.GetModules().Filter(func(p MigrationProvider) {
		for _, migration := range p.ProvideMigration() {
			if migration.Connection == connection {
				migrations.Collection = append(migrations.Collection, migration)
			}
		}
	})
	migrations.Db, _ = m.maker.Make(connection)
	return migrations
}

func (m Module) collectSeeds(connection string) Seeds {
	if connection == "" {
		connection = "default"
	}
	var seeds Seeds
	m.container.GetModules().Filter(func(p SeedProvider) {
		for _, seed := range p.ProvideSeed() {
			if seed.Connection == connection {
				seeds.Collection = append(seeds.Collection, seed)
			}
		}
	})
	seeds.Logger = m.logger
	seeds.Db, _ = m.maker.Make(connection)
	return seeds
}

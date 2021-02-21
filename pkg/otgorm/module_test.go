package otgorm

import (
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"testing"
)

type Mock struct {
	action string
}

func (m *Mock) ProvideSeed() []*Seed {
	return []*Seed{
		{
			Name:       "whatever",
			Connection: "default",
			Run: func(db *gorm.DB) error {
				m.action = "seeding"
				return nil
			},
		},
	}
}

func (m *Mock) ProvideMigration() []*Migration {
	return []*Migration{
		{
			ID:         "202101011000",
			Connection: "default",
			Migrate: func(db *gorm.DB) error {
				m.action = "migration"
				return nil
			},
			Rollback: func(db *gorm.DB) error {
				m.action = "rollback"
				return nil
			},
		},
	}
}

func TestModule_ProvideCommand(t *testing.T) {
	c := core.New(core.WithInline("gorm.default.database", "sqlite"),
		core.WithInline("gorm.default.dsn", "file::memory:?cache=shared"))
	c.ProvideEssentials()
	c.Provide(Provide)
	c.AddModuleFunc(New)
	mock := &Mock{}
	c.AddModule(mock)
	rootCmd := cobra.Command{}
	c.ApplyRootCommand(&rootCmd)
	assert.True(t, rootCmd.HasSubCommands())

	cases := []struct {
		name   string
		args   []string
		expect string
	}{
		{
			"migrate",
			[]string{"database", "migrate"},
			"migration",
		},
		{
			"rollback",
			[]string{"database", "migrate", "--rollback"},
			"rollback",
		},
		{
			"seed",
			[]string{"database", "seed"},
			"seeding",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			rootCmd.SetArgs(cc.args)
			rootCmd.Execute()
			assert.Equal(t, cc.expect, mock.action)
		})
	}
}

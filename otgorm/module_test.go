package otgorm

import (
	"context"
	"database/sql"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	mock_metrics "github.com/DoNewsCode/core/otgorm/mocks"
	"github.com/golang/mock/gomock"
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
	c.Provide(di.Deps{provideDatabaseFactory})
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

func TestModule_ProvideRunGroup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	idle := mock_metrics.NewMockGauge(ctrl)
	idle.EXPECT().With(gomock.Eq("default"), gomock.Eq("sqlite")).Return(idle).MinTimes(1)
	idle.EXPECT().Set(gomock.Eq(0.0)).MinTimes(1)

	open := mock_metrics.NewMockGauge(ctrl)
	open.EXPECT().With(gomock.Eq("default"), gomock.Eq("sqlite")).Return(open).MinTimes(1)
	open.EXPECT().Set(gomock.Eq(2.0)).MinTimes(1)

	inUse := mock_metrics.NewMockGauge(ctrl)
	inUse.EXPECT().With(gomock.Eq("default"), gomock.Eq("sqlite")).Return(inUse).MinTimes(1)
	inUse.EXPECT().Set(gomock.Eq(2.0)).MinTimes(1)

	c := core.New(
		core.WithInline("gorm.default.database", "sqlite"),
		core.WithInline("gorm.default.dsn", "file::memory:?cache=shared"),
		core.WithInline("gormMetrics.interval", "1ms"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(di.Deps{func() *Gauges {
		return &Gauges{
			Idle:  idle,
			InUse: inUse,
			Open:  open,
		}
	}})
	c.Provide(Providers())
	c.AddModuleFunc(New)

	ctx, cancel := context.WithCancel(context.Background())
	var (
		c1 *sql.Conn
		c2 *sql.Conn
	)
	c.Invoke(func(db *gorm.DB) {
		rawSQL, _ := db.DB()
		c1, _ = rawSQL.Conn(ctx)
		c2, _ = rawSQL.Conn(ctx)
	})
	go c.Serve(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
	c1.Close()
	c2.Close()
}

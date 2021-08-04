package mysqlstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/dtx"
	"github.com/DoNewsCode/core/dtx/sagas"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type module struct{}

func (m module) ProvideMigration() []*otgorm.Migration {
	return Migrations("default")
}

func TestMain(m *testing.M) {
	if os.Getenv("MYSQL_DSN") == "" {
		fmt.Println("Set env MYSQL_DSN to run mysqlstore tests")
		os.Exit(0)
	}
	c := core.New(
		core.WithInline("log.level", "error"),
		core.WithInline("gorm.default.database", "mysql"),
		core.WithInline("gorm.default.dsn", os.Getenv("MYSQL_DSN")),
	)
	c.ProvideEssentials()
	c.Provide(otgorm.Providers())
	c.AddModule(&module{})
	c.AddModuleFunc(otgorm.New)
	rootCmd := cobra.Command{}
	c.ApplyRootCommand(&rootCmd)
	rootCmd.SetArgs([]string{"database", "migrate"})
	rootCmd.Execute()
	code := m.Run()
	rootCmd.SetArgs([]string{"database", "migrate", "--rollback"})
	rootCmd.Execute()
	os.Exit(code)
}

func TestMySQLStore(t *testing.T) {
	cases := []struct {
		name string
		test func(db *gorm.DB)
	}{
		{
			"session",
			func(db *gorm.DB) {
				store := MySQLStore{db: db}
				err := store.Log(context.Background(), sagas.Log{
					ID:         "1",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Session,
					StepParam:  nil,
					StepName:   "test",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Ack(context.Background(), "1", nil)
				assert.NoError(t, err)
			},
		},
		{
			"do",
			func(db *gorm.DB) {
				store := MySQLStore{db: db}
				ctx := context.WithValue(context.Background(), dtx.CorrelationID, "foo")
				err := store.Log(ctx, sagas.Log{
					ID:         "2",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Session,
					StepParam:  nil,
					StepName:   "test",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "3",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Do,
					StepParam:  nil,
					StepName:   "first",
					StepError:  nil,
				})
				assert.NoError(t, err)
				logs, err := store.UnacknowledgedSteps(ctx, "foo")
				assert.NoError(t, err)
				assert.Len(t, logs, 1)
			},
		},
		{
			"undo",
			func(db *gorm.DB) {
				store := MySQLStore{db: db}
				ctx := context.WithValue(context.Background(), dtx.CorrelationID, "foo")
				err := store.Log(ctx, sagas.Log{
					ID:         "2",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Session,
					StepParam:  nil,
					StepName:   "test",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "3",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Do,
					StepParam:  nil,
					StepName:   "first",
					StepError:  nil,
				})
				assert.NoError(t, err)
				_ = store.Ack(ctx, "3", errors.New("foo error"))
				err = store.Log(ctx, sagas.Log{
					ID:         "4",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Undo,
					StepParam:  nil,
					StepName:   "first;",
					StepError:  nil,
				})
				assert.NoError(t, err)
				logs, err := store.UnacknowledgedSteps(ctx, "foo")
				assert.NoError(t, err)
				assert.Len(t, logs, 1)
			},
		},
		{
			"undo",
			func(db *gorm.DB) {
				store := MySQLStore{db: db}
				ctx := context.WithValue(context.Background(), dtx.CorrelationID, "foo")
				err := store.Log(ctx, sagas.Log{
					ID:         "2",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Session,
					StepParam:  nil,
					StepName:   "test",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "3",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Do,
					StepParam:  nil,
					StepName:   "first",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Ack(ctx, "3", errors.New("foo error"))
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "4",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Undo,
					StepParam:  nil,
					StepName:   "first",
					StepError:  nil,
				})
				assert.NoError(t, err)
				_ = store.Ack(ctx, "4", nil)
				logs, err := store.UnacknowledgedSteps(ctx, "foo")
				assert.NoError(t, err)
				assert.Len(t, logs, 0)
			},
		},
		{
			"session",
			func(db *gorm.DB) {
				store := MySQLStore{db: db}
				ctx := context.WithValue(context.Background(), dtx.CorrelationID, "session")
				err := store.Log(ctx, sagas.Log{
					ID:         "5",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Session,
					StepParam:  nil,
					StepName:   "test",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "6",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Do,
					StepParam:  nil,
					StepName:   "first",
					StepError:  nil,
				})
				assert.NoError(t, err)
				_ = store.Ack(ctx, "3", errors.New("foo error"))
				logs, err := store.UncommittedSagas(ctx)
				assert.NoError(t, err)
				assert.Len(t, logs, 1)
			},
		},
		{
			"multiple steps",
			func(db *gorm.DB) {
				store := MySQLStore{db: db}
				ctx := context.WithValue(context.Background(), dtx.CorrelationID, "multiple steps")
				err := store.Log(ctx, sagas.Log{
					ID:         "7",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Session,
					StepParam:  nil,
					StepName:   "test",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "8",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Do,
					StepParam:  nil,
					StepName:   "first",
					StepError:  nil,
				})
				assert.NoError(t, err)
				err = store.Ack(ctx, "8", nil)
				assert.NoError(t, err)
				err = store.Log(ctx, sagas.Log{
					ID:         "9",
					StartedAt:  time.Now(),
					FinishedAt: time.Time{},
					LogType:    sagas.Do,
					StepParam:  []byte(`foo second`),
					StepName:   "Second",
					StepError:  nil,
				})
				assert.NoError(t, err)
				_ = store.Ack(ctx, "9", errors.New("bar"))
				logs, err := store.UnacknowledgedSteps(ctx, "multiple steps")
				assert.NoError(t, err)
				assert.Len(t, logs, 2)
			},
		},
	}
	c := core.New(
		core.WithInline("log.level", "debug"),
		core.WithInline("gorm.default.database", "mysql"),
		core.WithInline("gorm.default.dsn", os.Getenv("MYSQL_DSN")),
	)
	c.ProvideEssentials()
	c.Provide(otgorm.Providers())
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			c.Invoke(cc.test)
		})
	}
}

func TestStore_CleanUp(t *testing.T) {
	c := core.New(
		core.WithInline("log.level", "error"),
		core.WithInline("gorm.default.database", "mysql"),
		core.WithInline("gorm.default.dsn", os.Getenv("MYSQL_DSN")),
	)
	c.ProvideEssentials()
	c.Provide(otgorm.Providers())
	c.Invoke(func(db *gorm.DB) {
		db.Exec("truncate table saga_logs")
		store := New(db, WithRetention(2*time.Hour), WithCleanUpInterval(time.Millisecond))

		store.Log(context.Background(), sagas.Log{
			ID:            "100",
			CorrelationID: "111",
			StartedAt:     time.Now().Add(-3 * time.Hour),
			LogType:       0,
			StepParam:     nil,
			StepName:      "test",
			StepError:     nil,
		})
		store.Log(context.Background(), sagas.Log{
			ID:            "100",
			CorrelationID: "112",
			StartedAt:     time.Now().Add(-1 * time.Hour),
			LogType:       0,
			StepParam:     nil,
			StepName:      "test",
			StepError:     nil,
		})
		store.cleanUp(context.Background())
		var count int64
		db.Table("saga_logs").Count(&count)
		assert.Equal(t, int64(1), count)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()
		assert.Error(t, store.CleanUp(ctx))
	})
}

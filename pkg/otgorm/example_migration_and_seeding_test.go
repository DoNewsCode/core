package otgorm_test

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/otgorm"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserName string
}

type Module struct{}

func (m Module) ProvideMigration() []*otgorm.Migration {
	return []*otgorm.Migration{
		{
			ID: "202010280100",
			Migrate: func(db *gorm.DB) error {
				type User struct {
					gorm.Model
					UserName string
				}
				return db.AutoMigrate(
					&User{},
				)
			},
			Rollback: func(db *gorm.DB) error {
				type User struct{}
				return db.Migrator().DropTable(&User{})
			},
		},
	}
}

func (m Module) ProvideSeed() []*otgorm.Seed {
	return []*otgorm.Seed{
		{
			ID:   "202010280200",
			Name: "seeding mysql",
			Run: func(db *gorm.DB) error {
				for i := 0; i < 100; i++ {
					db.Create(&User{
						UserName: "foo",
					})
				}
				return nil
			},
		},
	}
}

func Example_migrationAndSeeding() {
	c := core.New(
		core.WithInline("log.level", "error"),
		core.WithInline("gorm.default.database", "sqlite"),
		core.WithInline("gorm.default.dsn", "file::memory:?cache=shared"),
	)
	c.ProvideEssentials()
	c.Provide(otgorm.Provide)
	c.AddModule(&Module{})
	c.AddModuleFunc(otgorm.New)
	rootCmd := cobra.Command{}
	c.ApplyRootCommand(&rootCmd)
	rootCmd.SetArgs([]string{"database", "migrate"})
	rootCmd.Execute()
	rootCmd.SetArgs([]string{"database", "seed"})
	rootCmd.Execute()
	c.Invoke(func(db *gorm.DB) {
		var user User
		db.Last(&user)
		fmt.Println(user.UserName)
	})
	// Output:
	// foo
}

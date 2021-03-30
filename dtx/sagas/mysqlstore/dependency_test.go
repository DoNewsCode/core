package mysqlstore

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/dtx/sagas"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

func TestProviders(t *testing.T) {
	c := core.Default(
		core.WithInline("sagas-mysql.connection", "default"),
		core.WithInline("sagas-mysql.retention", "1h"),
		core.WithInline("sagas-mysql.cleanupInterval", "1h"),
	)
	c.Provide(otgorm.Providers())
	c.Provide(sagas.Providers())
	c.Provide(Providers())
	c.Invoke(func(r *sagas.Registry) {
		if _, ok := r.Store.(*MySQLStore); !ok {
			t.Fatal("r.Store should be a mysql store")
		}
	})
	c.Invoke(func(db *gorm.DB) {
		otgorm.Migrations{
			Db:         db,
			Collection: Migrations("default"),
		}.Migrate()
		otgorm.Migrations{
			Db:         db,
			Collection: Migrations("default"),
		}.Rollback("-1")
	})
}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	_, err := yaml.Marshal(conf.Config)
	assert.NoError(t, err)
}

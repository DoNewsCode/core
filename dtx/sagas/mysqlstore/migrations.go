package mysqlstore

import (
	"github.com/DoNewsCode/core/otgorm"
	"gorm.io/gorm"
)

// Migrations returns the database migrations needed for MySQLStore.
func Migrations() []*otgorm.Migration {
	return []*otgorm.Migration{
		{
			ID: "202103150100",
			Migrate: func(db *gorm.DB) error {
				return db.Exec("CREATE TABLE saga_logs (id blob, correlation_id varchar(255), started_at datetime, finished_at datetime, log_type smallint, step_name varchar(255), step_param blob, step_error varchar(255));").Error
			},
			Rollback: func(db *gorm.DB) error {
				type Log struct{}
				return db.Migrator().DropTable(&Log{})
			},
		},
	}
}

package otgorm

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Drivers is a map of string names and gorm.Dialector constructors. Inject Drivers to DI container to customize dialectors.
type Drivers map[string]func(dsn string) gorm.Dialector

func getDefaultDrivers() Drivers {
	return map[string]func(dsn string) gorm.Dialector{
		"mysql":      mysql.Open,
		"sqlite":     sqlite.Open,
		"clickhouse": clickhouse.Open,
	}
}

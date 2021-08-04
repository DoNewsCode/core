package otgorm

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Drivers map[string]func(dsn string) gorm.Dialector

func getDefaultDrivers() Drivers {
	return map[string]func(dsn string) gorm.Dialector{
		"mysql":      mysql.Open,
		"sqlite":     sqlite.Open,
		"clickhouse": clickhouse.Open,
	}
}

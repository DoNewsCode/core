package otgorm

import "gorm.io/gorm"

// GormConfigInterceptor is a function that allows user to Make last minute
// change to *gorm.Config when constructing *gorm.DB.
type GormConfigInterceptor func(name string, conf *gorm.Config)

type providersOption struct {
	interceptor GormConfigInterceptor
	drivers     Drivers
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithConfigInterceptor instructs the Providers to accept the
// GormConfigInterceptor so that users can change config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithConfigInterceptor(interceptor GormConfigInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.interceptor = interceptor
	}
}

// WithDriver instructs the Providers to add new drivers or replace existing
// drivers. (By default, only "mysql", "sqlite", and "clickhouse" drivers are
// registered out of box.) For example, if SqlServer driver and mysql driver are
// need, we can pass in the following option
// 	WithDriver(map[string]func(dsn string){"sqlserver": sqlServer.Open, "mysql": mysql.Open})
func WithDrivers(drivers Drivers) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.drivers = drivers
	}
}

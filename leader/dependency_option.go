package leader

type providersOption struct {
	driver            Driver
	driverConstructor func(args DriverConstructorArgs) (Driver, error)
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithDriver instructs the Providers to accept a leader election driver
// different from the default one. This option supersedes the
// WithDriverConstructor option.
func WithDriver(driver Driver) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.driver = driver
	}
}

// WithClientConstructor instructs the Providers to accept an alternative constructor for election driver.
// If the WithDriver option is set, this option becomes an no-op.
func WithDriverConstructor(f func(args DriverConstructorArgs) (Driver, error)) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.driverConstructor = f
	}
}

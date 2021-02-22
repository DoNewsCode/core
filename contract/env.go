package contract

// Env is the interface for environment of the application.
type Env interface {
	IsLocal() bool
	IsDevelopment() bool
	IsTesting() bool
	IsProduction() bool
	String() string
}

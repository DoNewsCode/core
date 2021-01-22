package contract

type Env interface {
	IsLocal() bool
	IsDevelopment() bool
	IsTesting() bool
	IsProduction() bool
	String() string
}

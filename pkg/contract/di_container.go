package contract

// DiContainer is a container roughly modeled after dig.Container
type DiContainer interface {
	Provide(constructor interface{}) error
	Invoke(function interface{}) error
}

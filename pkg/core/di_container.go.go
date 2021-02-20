package core

// diContainer is a container roughly modeled after dig.Container
type diContainer interface {
	Provide(constructor interface{}) error
	Invoke(function interface{}) error
}

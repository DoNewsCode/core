package config

type AppName string

func (a AppName) String() string {
	return string(a)
}

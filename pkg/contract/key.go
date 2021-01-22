package contract

type Keyer interface {
	Key(args ...string) string
}

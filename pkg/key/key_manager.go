package key

import (
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
)

type KeyManager struct {
	Prefixes []string
}

func NewManager(parts ...string) KeyManager {
	return KeyManager{
		Prefixes: parts,
	}
}

func (k KeyManager) Key(delimiter string, parts ...string) string {
	parts = append(k.Prefixes, parts...)
	return strings.Join(parts, delimiter)
}

func (k KeyManager) Spread() []string {
	return k.Prefixes
}

func (k KeyManager) With(parts ...string) KeyManager {
	newKeyManager := KeyManager{}
	newKeyManager.Prefixes = append(k.Prefixes, parts...)
	return newKeyManager
}

func With(k contract.Keyer, parts ...string) KeyManager {
	km := KeyManager{}
	parts = append(k.Spread(), parts...)
	return km.With(parts...)
}

func SpreadInterface(k contract.Keyer) []interface{} {
	var spreader = k.Spread()
	var out = make([]interface{}, len(spreader), len(spreader))
	for i := range k.Spread() {
		out[i] = interface{}(spreader[i])
	}
	return out
}

func KeepOdd(k contract.Keyer) contract.Keyer {
	var (
		spreader = k.Spread()
		km       = KeyManager{}
	)
	for i := range spreader {
		if i%2 == 1 {
			km.Prefixes = append(km.Prefixes, spreader[i])
		}
	}
	km.Prefixes = k.Spread()
	return km
}

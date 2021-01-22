package key

import (
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
)

type KeyManager struct {
	Prefixes  []string
	Delimiter string
}

func NewKeyManager(delimiter string, parts ...string) KeyManager {
	return KeyManager{
		Prefixes:  parts,
		Delimiter: delimiter,
	}
}

func (k KeyManager) Key(parts ...string) string {
	parts = append(k.Prefixes, parts...)
	return strings.Join(parts, k.Delimiter)
}
func (k KeyManager) With(parts ...string) KeyManager {
	newKeyManager := KeyManager{Delimiter: k.Delimiter}
	newKeyManager.Prefixes = append(k.Prefixes, parts...)
	return newKeyManager
}

func With(k contract.Keyer, parts ...string) KeyManager {
	del := ":"
	if kk, ok := k.(KeyManager); ok {
		del = kk.Delimiter
	}
	km := KeyManager{Delimiter: del}
	parts = append([]string{k.Key()}, parts...)
	return km.With(parts...)
}

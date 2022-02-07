/*
Package key provides a way to distribute labels to component.

Introduction

It is common that certain component requires a prefix to identify its owner. For example,
In redis we use labels to prefix the keys used. In metrics,
the labels are used to category metrics dimension. In logs, contextual information are
passed around as such keys.

Here is an example from go kit log, go kit metrics and go-redis:

	var (
		logger log.Logger
		counter metrics.Counter
		client redis.Client
	)
	log.With(logger, "module", "foo").Log("foo", "bar")
	client.Set("foo:mykey").Result()
	counter.With("module", "foo").Add(1)


Package key allows it to be rewritten as:

	keyer := key.New("module", "foo")
	logger := log.With(logger, keyer.Spread()...)
	client.Set(key.KeepOdd(keyer).Key(":", "mykey")).Result()
	counter.With(key.Spread()...).Add(1)

You don't need package key if such labels are simple and clustered in one place.
It is most beneficial if labels are used multiple times and are scattered all
over the place.

manager is immutable, hence safe for concurrent access.
*/
package key

import (
	"strings"

	"github.com/DoNewsCode/core/contract"
)

// manager is an immutable struct that manages the labels for log, metrics,
// tracing, kv store, etc.
type manager struct {
	Prefixes []string
}

// New constructs a manager from alternating key values.
//
//  manager := New("module", "foo", "service", "bar")
func New(parts ...string) manager {
	return manager{
		Prefixes: parts,
	}
}

// Key creates a string key composed by labels stored in manager
func (k manager) Key(delimiter string, parts ...string) string {
	parts = append(k.Prefixes, parts...)
	return strings.Join(parts, delimiter)
}

// Spread returns all labels in manager as []string.
func (k manager) Spread() []string {
	return k.Prefixes
}

// With returns a new manager with added alternating key values.
// Note: manager is immutable. With Creates a new instance.
func (k manager) With(parts ...string) manager {
	newKeyManager := manager{}
	newKeyManager.Prefixes = append(k.Prefixes, parts...)
	return newKeyManager
}

// With returns a new manager with added alternating key values.
// Note: manager is immutable. With Creates a new instance.
func With(k contract.Keyer, parts ...string) manager {
	km := manager{}
	parts = append(k.Spread(), parts...)
	return km.With(parts...)
}

// SpreadInterface likes Spread, but returns a slice of any
func SpreadInterface(k contract.Keyer) []any {
	spreader := k.Spread()
	out := make([]any, len(spreader))
	for i := range k.Spread() {
		out[i] = any(spreader[i])
	}
	return out
}

// KeepOdd only retains the odd values in the contract.Keyer. Note: The
// alternating key-values count from zero. Odd values are the "value" in
// key-value pairs. To avoid confusion, the KeepEven method is intentionally not
// provided.
func KeepOdd(k contract.Keyer) contract.Keyer {
	var (
		spreader = k.Spread()
		km       = manager{}
	)
	for i := range spreader {
		if i%2 == 1 {
			km.Prefixes = append(km.Prefixes, spreader[i])
		}
	}
	return km
}

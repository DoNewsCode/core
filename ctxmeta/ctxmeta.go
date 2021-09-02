// Package ctxmeta provides a helper type for request-scoped metadata. This
// package is inspired by https://github.com/peterbourgon/ctxdata. (License:
// https://github.com/peterbourgon/ctxdata/blob/master/LICENSE) The original
// package doesn't support collecting different groups of contextual data. This
// forked version allows it.
package ctxmeta

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/DoNewsCode/core/contract"
)

var _ contract.ConfigUnmarshaler = (*Baggage)(nil)

// KeyVal combines a string key with its abstract value into a single tuple.
// It's used internally, and as a return type for Slice.
type KeyVal struct {
	Key string
	Val interface{}
}

// ErrNoBaggage is returned by accessor methods when they're called on a nil
// pointer receiver. This typically means From was called on a context that
// didn't have Baggage injected into it previously via Inject.
var ErrNoBaggage = errors.New("no baggage in context")

// ErrIncompatibleType is returned by GetAs/Unmarshal if the value associated with a key
// isn't assignable to the provided target.
var ErrIncompatibleType = errors.New("incompatible type")

// ErrNotFound is returned by Get or other accessors when the key isn't present.
var ErrNotFound = errors.New("key not found")

// Baggage is an opaque type that can be injected into a context at e.g. the start
// of a request, updated with metadata over the course of the request, and then
// queried at the end of the request.
//
// When a new request arrives in your program, HTTP server, etc., use the New
// constructor with the incoming request's context to construct a new, empty
// Baggage. Use the returned context for all further operations on that request.
// Use the From helper function to retrieve a previously-injected Baggage from a
// context, and set or get metadata. At the end of the request, all metadata
// collected will be available from any point in the callstack.
type Baggage struct {
	c chan []KeyVal
}

// Unmarshal get the value at given path, and store it into the target variable. Target must
// be a pointer to an assignable type. Get will return ErrNotFound if the key
// is not found, and ErrIncompatibleType if the found value is not assignable to
// target. Unmarshal also implements contract.ConfigUnmarshaler.
func (b *Baggage) Unmarshal(path string, target interface{}) error {
	val, err := b.Get(path)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(target)
	t := v.Type()
	if t.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	targetType := t.Elem()
	if !reflect.TypeOf(val).AssignableTo(targetType) {
		return ErrIncompatibleType
	}

	v.Elem().Set(reflect.ValueOf(val))
	return nil
}

// Get the value associated with key, or return ErrNotFound. If this method is
// called on a nil Baggage pointer, it returns ErrNoBaggage.
func (b *Baggage) Get(key string) (value interface{}, err error) {
	if b == nil {
		return nil, ErrNoBaggage
	}

	s := <-b.c
	defer func() { b.c <- s }()

	for _, kv := range s {
		if kv.Key == key {
			return kv.Val, nil
		}
	}

	return nil, ErrNotFound
}

// Set key to value. If key already exists, it will be overwritten. If this method
// is called on a nil Baggage pointer, it returns ErrNoBaggage.
func (b *Baggage) Set(key string, value interface{}) (err error) {
	if b == nil {
		return ErrNoBaggage
	}

	s := <-b.c
	defer func() { b.c <- s }()

	for i := range s {
		if s[i].Key == key {
			s[i].Val = value
			s = append(s[:i], append(s[i+1:], s[i])...)
			return nil
		}
	}

	s = append(s, KeyVal{key, value})

	return nil
}

// Update key to the value returned from the callback. If key doesn't exist, it
// returns ErrNotFound. If this method is called on a nil Baggage pointer, it
// returns ErrNoBaggage.
func (b *Baggage) Update(key string, callback func(value interface{}) interface{}) (err error) {
	if b == nil {
		return ErrNoBaggage
	}

	s := <-b.c
	defer func() { b.c <- s }()

	for i := range s {
		if s[i].Key == key {
			s[i].Val = callback(s[i].Val)
			return nil
		}
	}

	return ErrNotFound
}

// Delete key from baggage. If key doesn't exist, it returns ErrNotFound. If the
// MetadataSet is not associated with an initialized baggage, it returns
// ErrNoBaggage.
func (b *Baggage) Delete(key interface{}) (err error) {
	if b == nil {
		return ErrNoBaggage
	}
	s := <-b.c
	defer func() { b.c <- s }()

	for i := range s {
		if s[i].Key == key {
			s = append(s[:i], s[i+1:]...)
			return nil
		}
	}

	return ErrNotFound
}

// Slice returns a slice of key/value pairs in the order in which they were set.
func (b *Baggage) Slice() []KeyVal {
	s := <-b.c
	defer func() { b.c <- s }()

	r := make([]KeyVal, len(s))
	copy(r, s)
	return r
}

// Map returns a map of key to value.
func (b *Baggage) Map() map[string]interface{} {
	s := <-b.c
	defer func() { b.c <- s }()

	mp := make(map[string]interface{}, len(s))
	for _, kv := range s {
		mp[kv.Key] = kv.Val
	}
	return mp
}

// MetadataSet is a group key to the contextual data stored the context.
// The data stored with different MetadataSet instances are not shared.
type MetadataSet struct {
	key *struct{}
}

// DefaultMetadata contains the default key for Baggage in the context. Use this if there
// is no need to categorize metadata, ie. put all data in one baggage.
var DefaultMetadata = MetadataSet{key: &struct{}{}}

// New constructs a new set of metadata. This metadata can be used to retrieve a group of contextual data.
// The data stored with different MetadataSet instances are not shared.
func New() *MetadataSet {
	return &MetadataSet{key: &struct{}{}}
}

// Inject constructs a Baggage object and injects it into the provided context
// under the context key determined the metadata instance. Use the returned
// context for all further operations. The returned Baggage can be queried at any
// point for metadata collected over the life of the context.
func (m *MetadataSet) Inject(ctx context.Context) (*Baggage, context.Context) {
	c := make(chan []KeyVal, 1)
	c <- make([]KeyVal, 0, 32)
	d := &Baggage{c: c}
	return d, context.WithValue(ctx, m.key, d)
}

// GetBaggage returns the Baggage stored in the context.
func (m *MetadataSet) GetBaggage(ctx context.Context) *Baggage {
	if val, ok := ctx.Value(m.key).(*Baggage); ok {
		return val
	}
	return nil
}

// Inject constructs a Baggage object and injects it into the provided context
// under the default context key. Use the returned context for all further
// operations. The returned Data can be queried at any point for metadata
// collected over the life of the context.
func Inject(ctx context.Context) (*Baggage, context.Context) {
	return DefaultMetadata.Inject(ctx)
}

// GetBaggage returns the default Baggage stored in the context.
func GetBaggage(ctx context.Context) *Baggage {
	return DefaultMetadata.GetBaggage(ctx)
}

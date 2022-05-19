package ots3

import "github.com/DoNewsCode/core/di"

// Maker is an interface for *Factory. Used as a type hint for injection.
type Maker interface {
	Make(name string) (*Manager, error)
}

// Factory can be used to connect to multiple s3 servers.
type Factory = di.Factory[*Manager]

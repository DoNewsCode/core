package otes

import (
	"github.com/DoNewsCode/core/di"
	"github.com/olivere/elastic/v7"
)

// Maker models Factory
type Maker interface {
	Make(name string) (*elastic.Client, error)
}

// Factory is a *di.Factory that creates *elastic.Client using a specific
// configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates *elastic.Client using a specific configuration entry.
func (r Factory) Make(name string) (*elastic.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

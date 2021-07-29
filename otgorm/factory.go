package otgorm

import (
	"github.com/DoNewsCode/core/di"
	"gorm.io/gorm"
)

// Factory is the *di.Factory that creates *gorm.DB under a specific
// configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates *gorm.DB under a specific configuration entry.
func (d Factory) Make(name string) (*gorm.DB, error) {
	db, err := d.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return db.(*gorm.DB), nil
}

// Maker models Factory
type Maker interface {
	Make(name string) (*gorm.DB, error)
}

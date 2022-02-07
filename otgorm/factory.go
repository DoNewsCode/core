package otgorm

import (
	"github.com/DoNewsCode/core/di"

	"gorm.io/gorm"
)

// Factory is the *di.Factory that creates *gorm.DB under a specific
// configuration entry.
type Factory = di.Factory[*gorm.DB]

// Maker models Factory
type Maker interface {
	Make(name string) (*gorm.DB, error)
}

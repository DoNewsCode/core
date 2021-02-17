package otgorm

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Seed is a action to populate the database with predefined values.
type Seed struct {
	Name       string
	Connection string
	Run        func(*gorm.DB) error
}

// Seeds is a collection of seed.
type Seeds struct {
	Logger     log.Logger
	Db         *gorm.DB
	Collection []*Seed
}

// Seed runs all the seeds collected by the application.
func (s *Seeds) Seed() error {
	for _, ss := range s.Collection {
		_ = level.Info(s.Logger).Log("msg", "seeding "+ss.Name)
		if err := ss.Run(s.Db); err != nil {
			return errors.Wrapf(err, "failed to seed %s", ss.Name)
		}
	}
	return nil
}

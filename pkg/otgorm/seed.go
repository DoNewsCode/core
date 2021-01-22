package otgorm

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Seed struct {
	Name string
	Run  func(*gorm.DB) error
}

type Seeds struct {
	logger log.Logger
	Db    *gorm.DB
	Seeds []Seed
}

func (s *Seeds) Seed() error {
	for _, ss := range s.Seeds {
		_ = level.Info(s.logger).Log("msg", "seeding "+ss.Name)
		if err := ss.Run(s.Db); err != nil {
			return errors.Wrapf(err, "failed to run %s", ss.Name)
		}
	}
	return nil
}

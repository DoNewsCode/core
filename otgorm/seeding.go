package otgorm

import (
	"sort"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Seed is a action to populate the database with predefined values.
type Seed struct {
	// ID is the sorting key. Usually a timestamp like "201601021504".
	ID string
	// Name is the human-readable identifier used in logs.
	Name string
	// Connection is the preferred database connection name, like "default".
	Connection string
	// Run is a function that seeds the database
	Run func(*gorm.DB) error
}

// Seeds is a collection of seed.
type Seeds struct {
	Logger     log.Logger
	Db         *gorm.DB
	Collection []*Seed
}

func (s *Seeds) Len() int {
	return len(s.Collection)
}

func (s *Seeds) Less(i, j int) bool {
	ivalue, _ := strconv.ParseInt(s.Collection[i].ID, 10, 64)
	jvalue, _ := strconv.ParseInt(s.Collection[j].ID, 10, 64)
	return ivalue < jvalue
}

func (s *Seeds) Swap(i, j int) {
	s.Collection[j], s.Collection[i] = s.Collection[i], s.Collection[j]
}

// Seed runs all the seeds collected by the application.
func (s *Seeds) Seed() error {
	sort.Sort(s)
	for _, ss := range s.Collection {
		if ss.Name == "" {
			ss.Name = ss.ID
		}
		_ = level.Info(s.Logger).Log("msg", "seeding "+ss.Name)
		if err := ss.Run(s.Db); err != nil {
			return errors.Wrapf(err, "failed to seed %s", ss.Name)
		}
	}
	return nil
}

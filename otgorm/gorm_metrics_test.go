package otgorm

import (
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otgorm/mocks"
	"github.com/golang/mock/gomock"
)

func TestCollector(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("set MYSQL_DSN to run TestCollector")
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	var g = NewGauges(m, m, m)
	m.EXPECT().With(gomock.Any()).MinTimes(3).Return(m)
	m.EXPECT().Set(gomock.Any()).Times(3)

	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(di.Deps{func() *Gauges { return g }})

	c.Invoke(func(factory Factory, g *Gauges) {
		factory.Make("default")
		c := newCollector(factory, g, time.Nanosecond)
		c.collectConnectionStats()
	})
}

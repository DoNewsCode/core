package otgorm

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otgorm/mocks"
	"github.com/go-kit/kit/metrics/generic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("set MYSQL_DSN to run TestCollector")
		return
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	g := NewGauges(m, m, m)
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

func TestObserve(t *testing.T) {
	foo := generic.NewGauge("foo")
	gauges := NewGauges(foo, foo, foo)
	gauges.Observe(sql.DBStats{})
	assert.ElementsMatch(t, gauges.idle.(*generic.Gauge).LabelValues(), []string{"dbname", "", "driver", ""})
}

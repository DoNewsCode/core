package observability

import (
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/DoNewsCode/core/otredis"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestProvideOpentracing(t *testing.T) {
	conf, _ := config.NewConfig(config.WithProviderLayer(rawbytes.Provider([]byte(sample)), yaml.Parser()))
	Out, cleanup, err := ProvideOpentracing(
		config.AppName("foo"),
		config.EnvTesting,
		ProvideJaegerLogAdapter(log.NewNopLogger()),
		conf,
	)
	assert.NoError(t, err)
	assert.NotNil(t, Out)
	cleanup()
}

func TestProvideHistogramMetrics(t *testing.T) {
	Out := ProvideHistogramMetrics(
		config.AppName("foo"),
		config.EnvTesting,
	)
	assert.NotNil(t, Out)
}

func TestProvideGORMMetrics(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(otgorm.Providers())
	c.Invoke(func(db *gorm.DB, g *otgorm.Gauges) {
		d, err := db.DB()
		if err != nil {
			t.Error(err)
		}
		stats := d.Stats()
		withValues := []string{"dbname", "default", "driver", db.Name()}
		g.Idle.
			With(withValues...).
			Set(float64(stats.Idle))

		g.InUse.
			With(withValues...).
			Set(float64(stats.InUse))

		g.Open.
			With(withValues...).
			Set(float64(stats.OpenConnections))
	})
}

func TestProvideRedisMetrics(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(otredis.Providers())
	c.Invoke(func(cli redis.UniversalClient, g *otredis.Gauges) {
		stats := cli.PoolStats()
		withValues := []string{"dbname", "default"}
		g.Hits.
			With(withValues...).
			Set(float64(stats.Hits))
		g.Misses.
			With(withValues...).
			Set(float64(stats.Misses))
		g.Timeouts.
			With(withValues...).
			Set(float64(stats.Timeouts))
		g.TotalConns.
			With(withValues...).
			Set(float64(stats.TotalConns))
		g.IdleConns.
			With(withValues...).
			Set(float64(stats.IdleConns))
		g.StaleConns.
			With(withValues...).
			Set(float64(stats.StaleConns))
	})
}

func Test_provideConfig(t *testing.T) {
	Conf := provideConfig()
	assert.NotEmpty(t, Conf.Config)
}

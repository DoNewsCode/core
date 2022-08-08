package pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/di"
)

func Providers() di.Deps {
	return di.Deps{
		provideDefaultPool,
		providePoolFactory(),
		di.Bind(new(*Factory), new(Maker)),
	}
}

func provideDefaultPool(maker Maker) (*Pool, error) {
	return maker.Make("default")
}

// factoryIn is the injection parameter for provideDatabaseOut.
type factoryIn struct {
	di.In

	Conf          contract.ConfigUnmarshaler
	OnReloadEvent lifecycle.ConfigReload `optional:"true"`
}

type poolConfig struct {
	Cap         int32           `yaml:"cap" json:"cap"`
	Concurrency int32           `yaml:"concurrency" json:"concurrency"`
	Timeout     config.Duration `yaml:"timeout" json:"timeout"`
}

// out
type out struct {
	di.Out

	Factory *Factory
}

func providePoolFactory() func(p factoryIn) (out, func(), error) {
	var wg = &sync.WaitGroup{}
	return func(factoryIn factoryIn) (out, func(), error) {
		factory := di.NewFactory[*Pool](func(name string) (pair di.Pair[*Pool], err error) {
			var (
				conf poolConfig
			)
			if err := factoryIn.Conf.Unmarshal(fmt.Sprintf("pool.%s", name), &conf); err != nil {
				if name != "default" {
					return pair, fmt.Errorf("pool configuration %s not valid: %w", name, err)
				}
			}
			pool := &Pool{
				cap:         10,
				concurrency: 1000,
				ch:          make(chan job),
				wg:          wg,
				timeout:     10 * time.Minute,
			}
			if conf.Cap > 0 {
				pool.cap = conf.Cap
			}
			if conf.Concurrency > 0 {
				pool.concurrency = conf.Concurrency
			}
			if !conf.Timeout.IsZero() {
				pool.timeout = conf.Timeout.Duration
			}

			return di.Pair[*Pool]{
				Conn:   pool,
				Closer: nil,
			}, err
		})

		return out{
			Factory: &Factory{
				factory: factory,
				wg:      wg,
			},
		}, factory.Close, nil
	}
}

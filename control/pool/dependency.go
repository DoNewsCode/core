package pool

import (
	"context"
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
	Counter       *Counter               `optional:"true"`
}

type poolConfig struct {
	// Cap is the maximum number of tasks. If it exceeds the maximum number, new workers will be added automatically.
	// Default is 10.
	Cap int32 `yaml:"cap" json:"cap"`
	// Concurrency limits the maximum number of workers.
	// Default is 1000.
	Concurrency int32 `yaml:"concurrency" json:"concurrency"`
	// IdleTimeout idle workers will be recycled after this duration.
	// Default is 10 minutes.
	IdleTimeout config.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

// out
type out struct {
	di.Out

	Factory *Factory
}

func providePoolFactory() func(p factoryIn) (out, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	var worker = &worker{
		ch:            make(chan job),
		incWorkerChan: make(chan int32),
		cap:           10,
		timeout:       10 * time.Minute,
	}
	return func(factoryIn factoryIn) (out, func(), error) {
		factory := di.NewFactory[*Pool](func(name string) (pair di.Pair[*Pool], err error) {

			pool := &Pool{
				ch:              worker.ch,
				incJobCountFunc: worker.incJobCount,
			}
			if factoryIn.Counter != nil {
				pool.counter = factoryIn.Counter.PoolName(name)
			}

			return di.Pair[*Pool]{
				Conn:   pool,
				Closer: nil,
			}, err
		})
		var (
			conf poolConfig
		)
		_ = factoryIn.Conf.Unmarshal("pool", &conf)
		if conf.Cap > 0 {
			worker.cap = conf.Cap
		}
		if conf.Concurrency > 0 {
			worker.concurrency = conf.Concurrency
		}
		if !conf.IdleTimeout.IsZero() {
			worker.timeout = conf.IdleTimeout.Duration
		}
		worker.run(ctx)
		return out{
				Factory: factory,
			}, func() {
				cancel()
				factory.Close()
			}, nil
	}
}

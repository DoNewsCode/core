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
		if !conf.Timeout.IsZero() {
			worker.timeout = conf.Timeout.Duration
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

// Package modqueue contains integration with package core
package modqueue

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/DoNewsCode/std/pkg/queue"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/oklog/run"
	"runtime"
	"time"
)

// Gauge is an alias used for dependency injection
type Gauge metrics.Gauge

// DispatcherIn is the injection parameters for ProvideDispatcher
type DispatcherIn struct {
	di.In

	Conf        contract.ConfigAccessor
	Dispatcher  contract.Dispatcher
	RedisClient redis.UniversalClient
	Logger      log.Logger
	AppName     contract.AppName
	Env         contract.Env
	Gauge       Gauge `optional:"true"`
}

// DispatcherOut is the dig output of ProvideDispatcher
type DispatcherOut struct {
	di.Out
	di.Module

	Dispatcher          queue.Dispatcher
	DispatcherMaker     queue.DispatcherMaker
	QueueableDispatcher *queue.QueueableDispatcher
	DispatcherFactory   *queue.DispatcherFactory
}

// ProvideDispatcher is a provider for *DispatcherFactory and *QueueableDispatcher.
// It also provides an extracted interface for each.
func ProvideDispatcher(p DispatcherIn) (DispatcherOut, error) {
	var (
		err        error
		queueConfs map[string]queue.Conf
	)
	err = p.Conf.Unmarshal("queue", &queueConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf queue.Conf
		)
		if conf, ok = queueConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("queue configuration %s not found", name)
		}
		if p.Gauge != nil {
			p.Gauge = p.Gauge.With("queue", name)
		}
		queuedDispatcher := queue.WithQueue(p.Dispatcher, &queue.RedisDriver{
			Logger:      p.Logger,
			RedisClient: p.RedisClient,
			ChannelConfig: queue.ChannelConfig{
				Delayed:  fmt.Sprintf("{%s:%s:%s}:delayed", p.AppName.String(), p.Env.String(), name),
				Failed:   fmt.Sprintf("{%s:%s:%s}:failed", p.AppName.String(), p.Env.String(), name),
				Reserved: fmt.Sprintf("{%s:%s:%s}:reserved", p.AppName.String(), p.Env.String(), name),
				Waiting:  fmt.Sprintf("{%s:%s:%s}:waiting", p.AppName.String(), p.Env.String(), name),
				Timeout:  fmt.Sprintf("{%s:%s:%s}:timeout", p.AppName.String(), p.Env.String(), name),
			},
		}, queue.UseLogger(p.Logger), queue.UseParallelism(conf.Parallelism), queue.UseGauge(
			p.Gauge,
			time.Duration(conf.CheckQueueLengthIntervalSecond)*time.Second,
		))
		return async.Pair{
			Closer: nil,
			Conn:   queuedDispatcher,
		}, nil
	})

	// QueueableDispatcher must be created eagerly, so that the consumer goroutines can start on boot up.
	for name := range queueConfs {
		factory.Make(name)
	}

	dispatcherFactory := &queue.DispatcherFactory{Factory: factory}
	defaultQueueableDispatcher, err := dispatcherFactory.Make("default")
	return DispatcherOut{
		QueueableDispatcher: defaultQueueableDispatcher,
		Dispatcher:          defaultQueueableDispatcher,
		DispatcherFactory:   dispatcherFactory,
		DispatcherMaker:     dispatcherFactory,
	}, nil
}

// ProvideRunGroup implements RunProvider.
func (s DispatcherOut) ProvideRunGroup(group *run.Group) {
	for name := range s.DispatcherFactory.List() {
		queueName := name
		ctx, cancel := context.WithCancel(context.Background())
		group.Add(func() error {
			consumer, err := s.DispatcherFactory.Make(queueName)
			if err != nil {
				return err
			}
			return consumer.Consume(ctx)
		}, func(err error) {
			cancel()
		})
	}
}

// ProvideRunGroup implements RunProvider.
func (s DispatcherOut) ProvideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{{
		Name: "queue",
		Data: map[string]interface{}{
			"queue": map[string]queue.Conf{
				"default": {
					Parallelism:                    runtime.NumCPU(),
					CheckQueueLengthIntervalSecond: 15,
				},
			},
		},
	}}
}

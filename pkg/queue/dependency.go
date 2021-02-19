package queue

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/oklog/run"
)

// Gauge is an alias used for dependency injection
type Gauge metrics.Gauge

// Dispatcher is the key of *QueueableDispatcher in the dependencies graph. Used as a type hint for injection.
type Dispatcher interface {
	contract.Dispatcher
	Consume(ctx context.Context) error
}

// DispatcherMaker is the key of *DispatcherFactory in the dependencies graph. Used as a type hint for injection.
type DispatcherMaker interface {
	Make(string) (*QueueableDispatcher, error)
}

type configuration struct {
	Parallelism                    int `yaml:"parallelism" json:"parallelism"`
	CheckQueueLengthIntervalSecond int `yaml:"checkQueueLengthIntervalSecond" json:"checkQueueLengthIntervalSecond"`
}

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

// DispatcherOut is the di output of ProvideDispatcher
type DispatcherOut struct {
	di.Out
	di.Module

	Dispatcher          Dispatcher
	DispatcherMaker     DispatcherMaker
	QueueableDispatcher *QueueableDispatcher
	DispatcherFactory   *DispatcherFactory
	ExportedConfig      []contract.ExportedConfig `group:"config,flatten"`
}

// ProvideDispatcher is a provider for *DispatcherFactory and *QueueableDispatcher.
// It also provides an interface for each.
func ProvideDispatcher(p DispatcherIn) (DispatcherOut, error) {
	var (
		err        error
		queueConfs map[string]configuration
	)
	err = p.Conf.Unmarshal("queue", &queueConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf configuration
		)
		if conf, ok = queueConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("queue configuration %s not found", name)
		}
		if p.Gauge != nil {
			p.Gauge = p.Gauge.With("queue", name)
		}
		redisDriver := &RedisDriver{
			Logger:      p.Logger,
			RedisClient: p.RedisClient,
			ChannelConfig: ChannelConfig{
				Delayed:  fmt.Sprintf("{%s:%s:%s}:delayed", p.AppName.String(), p.Env.String(), name),
				Failed:   fmt.Sprintf("{%s:%s:%s}:failed", p.AppName.String(), p.Env.String(), name),
				Reserved: fmt.Sprintf("{%s:%s:%s}:reserved", p.AppName.String(), p.Env.String(), name),
				Waiting:  fmt.Sprintf("{%s:%s:%s}:waiting", p.AppName.String(), p.Env.String(), name),
				Timeout:  fmt.Sprintf("{%s:%s:%s}:timeout", p.AppName.String(), p.Env.String(), name),
			},
		}
		queuedDispatcher := WithQueue(
			p.Dispatcher,
			redisDriver,
			UseLogger(p.Logger),
			UseParallelism(conf.Parallelism),
			UseGauge(p.Gauge, time.Duration(conf.CheckQueueLengthIntervalSecond)*time.Second),
		)
		return async.Pair{
			Closer: nil,
			Conn:   queuedDispatcher,
		}, nil
	})

	// QueueableDispatcher must be created eagerly, so that the consumer goroutines can start on boot up.
	for name := range queueConfs {
		factory.Make(name)
	}

	dispatcherFactory := &DispatcherFactory{Factory: factory}
	defaultQueueableDispatcher, _ := dispatcherFactory.Make("default")
	return DispatcherOut{
		QueueableDispatcher: defaultQueueableDispatcher,
		Dispatcher:          defaultQueueableDispatcher,
		DispatcherFactory:   dispatcherFactory,
		DispatcherMaker:     dispatcherFactory,
		ExportedConfig:      provideConfig(),
	}, nil
}

// ProvideRunGroup implements RunProvider.
func (d DispatcherOut) ProvideRunGroup(group *run.Group) {
	for name := range d.DispatcherFactory.List() {
		queueName := name
		ctx, cancel := context.WithCancel(context.Background())
		group.Add(func() error {
			consumer, err := d.DispatcherFactory.Make(queueName)
			if err != nil {
				return err
			}
			return consumer.Consume(ctx)
		}, func(err error) {
			cancel()
		})
	}
}

// DispatcherFactory is a factory for *QueueableDispatcher. Note DispatcherFactory doesn't contain the factory method
// itself. ie. How to factory a dispatcher left there for users to define. Users then can use this type to create
// their own dispatcher implementation.
//
// Here is an example on how to create a custom DispatcherFactory with an InProcessDriver.
//
//		factory := async.NewFactory(func(name string) (async.Pair, error) {
//			queuedDispatcher := queue.WithQueue(
//				&events.SyncDispatcher{},
//				queue.NewInProcessDriver(),
//			)
//			return async.Pair{Conn: queuedDispatcher}, nil
//		})
//		dispatcherFactory := DispatcherFactory{Factory: factory}
//
type DispatcherFactory struct {
	*async.Factory
}

// Make returns a QueueableDispatcher by the given name. If it has already been created under the same name,
// the that one will be returned.
func (s *DispatcherFactory) Make(name string) (*QueueableDispatcher, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*QueueableDispatcher), nil
}

func provideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{{
		Owner: "queue",
		Data: map[string]interface{}{
			"queue": map[string]configuration{
				"default": {
					Parallelism:                    runtime.NumCPU(),
					CheckQueueLengthIntervalSecond: 15,
				},
			},
		},
	}}
}

package queue

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/oklog/run"
	"go.uber.org/dig"
	"time"
)

// QueuedDispatcher is an alias of contract.Dispatcher. Inject it if persistent event feature is needed.
type QueuedDispatcher interface {
	contract.Dispatcher
	Consume(ctx context.Context) error
}

// QueuedDispatcherParam is the injection parameters for ProvideQueuedDispatcher
type QueuedDispatcherParam struct {
	dig.In

	Conf        contract.ConfigAccessor
	Dispatcher  contract.Dispatcher
	RedisClient redis.UniversalClient
	Logger      log.Logger
	AppName     contract.AppName
	Env         contract.Env
	Gauge       metrics.Gauge `optional:"true"`
}

// ProvideQueuedDispatcher is a provider for QueuedDispatcher
func ProvideQueuedDispatcher(p QueuedDispatcherParam) (QueuedDispatcher, error) {
	factory := ProvideQueuedDispatcherFactory(p)
	conn, err := factory.Make("default")
	return conn, err
}

// QueuedDispatcherFactory is a factory for ProvideQueuedDispatcher
type QueuedDispatcherFactory struct {
	contract.Module
	*async.Factory
}

// Make returns a QueuedDispatcher by the given name. If it has already been created under the same name,
// the that one will be returned.
func (s QueuedDispatcherFactory) Make(name string) (QueuedDispatcher, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(QueuedDispatcher), nil
}

// ProvideQueuedDispatcherFactory is a provider for QueuedDispatcherFactory
func ProvideQueuedDispatcherFactory(p QueuedDispatcherParam) QueuedDispatcherFactory {
	type queueConf struct {
		Parallelism                    int
		CheckQueueLengthIntervalSecond int
	}
	var (
		err        error
		queueConfs map[string]queueConf
	)
	err = p.Conf.Unmarshal("queue", &queueConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf queueConf
		)
		if conf, ok = queueConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("queue configuration %s not found", name)
		}
		queuedDispatcher := WithQueue(p.Dispatcher, &RedisDriver{
			Logger:      p.Logger,
			RedisClient: p.RedisClient,
			ChannelConfig: ChannelConfig{
				Delayed:  fmt.Sprintf("{%s:%s:%s}:delayed", p.AppName.String(), p.Env.String(), name),
				Failed:   fmt.Sprintf("{%s:%s:%s}:failed", p.AppName.String(), p.Env.String(), name),
				Reserved: fmt.Sprintf("{%s:%s:%s}:reserved", p.AppName.String(), p.Env.String(), name),
				Waiting:  fmt.Sprintf("{%s:%s:%s}:waiting", p.AppName.String(), p.Env.String(), name),
				Timeout:  fmt.Sprintf("{%s:%s:%s}:timeout", p.AppName.String(), p.Env.String(), name),
			},
		}, UseLogger(p.Logger), UseParallelism(conf.Parallelism))
		queuedDispatcher.queueLengthGauge = p.Gauge.With("queue", name)
		queuedDispatcher.checkQueueLengthInterval = time.Duration(conf.CheckQueueLengthIntervalSecond) * time.Second
		return async.Pair{
			Closer: nil,
			Conn:   queuedDispatcher,
		}, nil
	})
	for name := range queueConfs {
		factory.Make(name)
	}
	return QueuedDispatcherFactory{Factory: factory}
}

// ProvideRunGroup implements RunProvider.
func (s *QueuedDispatcherFactory) ProvideRunGroup(group *run.Group) {
	for name := range s.List() {
		ctx, cancel := context.WithCancel(context.Background())
		group.Add(func() error {
			consumer, err := s.Make(name)
			if err != nil {
				return err
			}
			return consumer.Consume(ctx)
		}, func(err error) {
			cancel()
		})
	}
}

package queue

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/oklog/run"
	"golang.org/x/sync/errgroup"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/event"
	"github.com/pkg/errors"
)

// persistent is an interface that describes a persisted event.
type persistent interface {
	Defer() time.Duration
	Decorate(s *PersistedEvent)
}

// dispatcher is an extension of event.dispatcher. It adds the persistent event feature to event.dispatcher.
type dispatcher struct {
	contract.Module

	logger                   log.Logger
	driver                   Driver
	packer                   Packer
	rwLock                   sync.RWMutex
	reflectTypes             map[string]reflect.Type
	base                     contract.Dispatcher
	parallelism              int
	queueLengthGauge         metrics.Gauge
	checkQueueLengthInterval time.Duration
}

func (d *dispatcher) Dispatch(ctx context.Context, e contract.Event) error {
	if _, ok := e.(*PersistedEvent); ok {
		rType := d.reflectType(e.Type())
		ptr := reflect.New(rType)
		err := d.packer.Decompress(e.Data().([]byte), ptr)
		if err != nil {
			return errors.Wrapf(err, "dispatch serialized %s failed", e.Type())
		}
		return d.base.Dispatch(ctx, event.NewEvent(ptr.Elem().Interface()))
	}
	if _, ok := e.(persistent); ok {
		data, err := d.packer.Compress(e.Data())
		if err != nil {
			return errors.Wrapf(err, "dispatch deferrable %s failed", e.Type())
		}
		msg := &PersistedEvent{
			Attempts: 1,
			Value:    data,
		}
		e.(persistent).Decorate(msg)
		return d.driver.Push(ctx, msg, e.(persistent).Defer())
	}
	return d.base.Dispatch(ctx, e)
}

func (d *dispatcher) Subscribe(listener contract.Listener) {
	d.rwLock.Lock()
	for _, e := range listener.Listen() {
		d.reflectTypes[e.Type()] = reflect.TypeOf(e.Data())
	}
	d.rwLock.Unlock()
	d.base.Subscribe(listener)
}

// Consume starts the runner and blocks until context canceled or error occurred.
func (d *dispatcher) Consume(ctx context.Context) error {
	var jobChan = make(chan *PersistedEvent)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer close(jobChan)
		for {
			msg, err := d.driver.Pop(ctx)
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return err
			}
			jobChan <- msg
		}
	})

	if d.queueLengthGauge != nil {
		if d.checkQueueLengthInterval == 0 {
			d.checkQueueLengthInterval = 15 * time.Second
		}
		ticker := time.NewTicker(d.checkQueueLengthInterval)
		g.Go(func() error {
			for {
				select {
				case <-ticker.C:
					d.gauge(ctx)
				case <-ctx.Done():
					ticker.Stop()
					return ctx.Err()
				}
			}
		})
	}

	for i := 0; i < d.parallelism; i++ {
		g.Go(func() error {
			for msg := range jobChan {
				d.work(ctx, msg)
			}
			return nil
		})
	}
	return g.Wait()
}

// ProvideRunGroup implements RunProvider.
func (d *dispatcher) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(func() error {
		return d.Consume(ctx)
	}, func(err error) {
		cancel()
	})
}

func (d *dispatcher) work(ctx context.Context, msg *PersistedEvent) {
	ctx, cancel := context.WithTimeout(ctx, msg.HandleTimeout)
	defer cancel()
	err := d.Dispatch(ctx, msg)
	if err != nil {
		if msg.Attempts < msg.MaxAttempts {
			_ = level.Info(d.logger).Log("err", errors.Wrapf(err, "event %s failed %d times, retrying", msg.Key, msg.Attempts))
			_ = d.Dispatch(context.Background(), event.NewEvent(RetryingEvent{Err: err, Msg: msg}))
			_ = d.driver.Retry(context.Background(), msg)
			return
		}
		_ = level.Warn(d.logger).Log("err", errors.Wrapf(err, "event %s failed after %d Attempts, aborted", msg.Key, msg.MaxAttempts))
		_ = d.Dispatch(context.Background(), event.NewEvent(AbortedEvent{Err: err, Msg: msg}))
		_ = d.driver.Fail(context.Background(), msg)
		return
	}
	_ = d.driver.Ack(context.Background(), msg)
}

func (d *dispatcher) reflectType(typeName string) reflect.Type {
	d.rwLock.RLock()
	defer d.rwLock.RUnlock()
	return d.reflectTypes[typeName]
}

func (d *dispatcher) gauge(ctx context.Context) {
	queueInfo, err := d.driver.Info(ctx)
	if err != nil {
		_ = level.Warn(d.logger).Log("err", err)
	}
	d.queueLengthGauge.With("queue", "failed").Set(float64(queueInfo.Failed))
	d.queueLengthGauge.With("queue", "delayed").Set(float64(queueInfo.Delayed))
	d.queueLengthGauge.With("queue", "timeout").Set(float64(queueInfo.Timeout))
	d.queueLengthGauge.With("queue", "waiting").Set(float64(queueInfo.Waiting))
}

// UsePacker allows consumer to replace the default Packer with a custom one. UsePacker is an option for WithQueue.
func UsePacker(packer Packer) func(*dispatcher) {
	return func(dispatcher *dispatcher) {
		dispatcher.packer = packer
	}
}

// UseLogger is an option for WithQueue that feeds the queue with a Logger of choice.
func UseLogger(logger log.Logger) func(*dispatcher) {
	return func(dispatcher *dispatcher) {
		dispatcher.logger = logger
	}
}

// UseParallelism is an option for WithQueue that sets the parallelism for queue consumption
func UseParallelism(parallelism int) func(*dispatcher) {
	return func(dispatcher *dispatcher) {
		dispatcher.parallelism = parallelism
	}
}

// UseGauge is an option for WithQueue that collects a gauge metrics
func UseGauge(gauge metrics.Gauge, interval time.Duration) func(*dispatcher) {
	return func(dispatcher *dispatcher) {
		dispatcher.queueLengthGauge = gauge
		dispatcher.checkQueueLengthInterval = interval
	}
}

// WithQueue wraps a dispatcher and returns a decorated dispatcher. The latter dispatcher now can send and
// listen to "persisted" events. Those persisted events will guarantee at least one execution, as they are stored in an
// external storage and won't be released until the dispatcher acknowledges the end of execution.
func WithQueue(baseDispatcher contract.Dispatcher, driver Driver, opts ...func(*dispatcher)) *dispatcher {
	qd := dispatcher{
		driver:       driver,
		packer:       packer{},
		rwLock:       sync.RWMutex{},
		reflectTypes: make(map[string]reflect.Type),
		base:         baseDispatcher,
		parallelism:  runtime.NumCPU(),
	}
	for _, f := range opts {
		f(&qd)
	}
	return &qd
}

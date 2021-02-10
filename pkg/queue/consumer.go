package queue

import (
	"context"
	"github.com/DoNewsCode/std/pkg/logging"
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// WithPacker is an option for NewConsumer that replace default Packer with a Packer of choice.
func WithPacker(packer Packer) func(consumer *Consumer) {
	return func(consumer *Consumer) {
		consumer.packer = packer
	}
}

// WithLogger is an option for NewConsumer that feeds the consumer with a logger of choice.
func WithLogger(logger log.Logger) func(consumer *Consumer) {
	return func(consumer *Consumer) {
		consumer.logger = logger
	}
}

// WithParallelism is an option for NewConsumer that limits the number of go routine workers running in parallel.
func WithParallelism(parallelism int) func(consumer *Consumer) {
	return func(consumer *Consumer) {
		consumer.parallelism = parallelism
	}
}

// NewConsumer creates a new instance of consumer
func NewConsumer(dispatcher contract.Dispatcher, driver Driver, opts ...func(consumer *Consumer)) *Consumer {
	c := &Consumer{
		packer:      packer{},
		logger:      logging.NewLogger("logfmt"),
		driver:      driver,
		dispatcher:  dispatcher,
		parallelism: 1,
	}
	for _, f := range opts {
		f(c)
	}
	return c
}

// Consumer defines a type of queue consumer.
type Consumer struct {
	packer      Packer
	logger      log.Logger
	driver      Driver
	dispatcher  contract.Dispatcher
	parallelism int
}

// Consume starts the runner and blocks until context canceled or error occurred.
func (c *Consumer) Consume(ctx context.Context) error {
	if _, ok := c.dispatcher.(*queueDispatcher); !ok {
		return errors.New("event dispatcher must be a queueDispatcher")
	}
	var jobChan = make(chan *SerializedMessage)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer close(jobChan)
		for {
			msg, err := c.driver.Pop(ctx)
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return err
			}
			jobChan <- msg
		}
	})

	for i := 0; i < c.parallelism; i++ {
		g.Go(func() error {
			for msg := range jobChan {
				c.work(ctx, msg)
			}
			return nil
		})
	}
	return g.Wait()
}

func (c *Consumer) work(ctx context.Context, msg *SerializedMessage) {
	ctx, cancel := context.WithTimeout(ctx, msg.HandleTimeout)
	defer cancel()
	err := c.dispatcher.Dispatch(ctx, msg)
	if err != nil {
		if msg.Attempts < msg.MaxAttempts {
			_ = level.Info(c.logger).Log("err", errors.Wrapf(err, "event %s failed %d times, retrying", msg.Key, msg.Attempts))
			_ = c.driver.Retry(context.Background(), msg)
			return
		}
		_ = level.Warn(c.logger).Log("err", errors.Wrapf(err, "event %s failed after %d Attempts, aborted", msg.Key, msg.MaxAttempts))
		_ = c.driver.Fail(context.Background(), msg)
		return
	}
	_ = c.driver.Ack(context.Background(), msg)
}

// Driver is the interface for queue engines. See RedisDriver for usage.
type Driver interface {
	// Push pushes the message onto the queue. It is possible to specify a time delay. If so the message
	// will be read after the delay. Use zero value if a delay is not needed.
	Push(ctx context.Context, message *SerializedMessage, delay time.Duration) error
	// Pop pops the message out of the queue. It blocks until a message is available or a timeout is reached.
	Pop(ctx context.Context) (*SerializedMessage, error)
	// Ack acknowledges a message has been processed.
	Ack(ctx context.Context, message *SerializedMessage) error
	// \Fail marks a message has failed.
	Fail(ctx context.Context, message *SerializedMessage) error
	// Reload put failed/timeout message back to the Waiting queue. If the temporary outage have been cleared,
	// messages can be tried again via Reload. Reload is not a normal retry.
	// It similarly gives otherwise dead messages one more chance,
	// but this chance is not subject to the limit of MaxAttempts, nor does it reset the number of time attempted.
	Reload(ctx context.Context, channel string) (int64, error)
	// Flush empties the queue under channel
	Flush(ctx context.Context, channel string) error
	// Info lists QueueInfo by inspecting queues one by one. Useful for metrics and monitor.
	Info(ctx context.Context) (QueueInfo, error)
	// Retry put the message back onto the delayed queue.
	Retry(ctx context.Context, message *SerializedMessage) error
}

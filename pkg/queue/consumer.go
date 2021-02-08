package queue

import (
	"context"
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Consumer struct {
	packer      Packer
	logger      log.Logger
	driver      Driver
	dispatcher  contract.Dispatcher
	parallelism int
}

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

type Driver interface {
	Push(ctx context.Context, message *SerializedMessage, delay time.Duration) error
	Pop(ctx context.Context) (*SerializedMessage, error)
	Ack(ctx context.Context, message *SerializedMessage) error
	Fail(ctx context.Context, message *SerializedMessage) error
	Reload(ctx context.Context, channel string) (int64, error)
	Flush(ctx context.Context, channel string) error
	Info(ctx context.Context) (QueueInfo, error)
	Retry(ctx context.Context, message *SerializedMessage) error
}

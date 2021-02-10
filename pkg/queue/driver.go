package queue

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// The Packer interface describes how to save the message in wire format
type Packer interface {
	// Compress serializes the message to bytes
	Compress(message interface{}) ([]byte, error)
	// Decompress reverses the bytes to message
	Decompress(data []byte, message interface{}) error
}

// RedisDriver is a queue driver backed by redis. It is easy to setup, and offers at least once semantic.
type RedisDriver struct {
	logger        log.Logger
	redisClient   redis.UniversalClient
	channelConfig ChannelConfig
	popTimeout    time.Duration
	packer        Packer
}

// Push pushes the message onto the queue. It is possible to specify a time delay. If so the message
// will be read after the delay. Use zero value if a delay is not needed.
func (r RedisDriver) Push(ctx context.Context, message *SerializedMessage, delay time.Duration) error {
	data, err := r.packer.Compress(message)
	if err != nil {
		return errors.Wrap(err, "failed to compress message")
	}
	if delay <= time.Duration(0) {
		_, err = r.redisClient.LPush(ctx, r.channelConfig.Waiting, data).Result()
		if err != nil {
			return errors.Wrap(err, "failed to lpush while pushing")
		}
		return nil
	}
	_, err = r.redisClient.ZAdd(ctx, r.channelConfig.Delayed, &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: data,
	}).Result()
	if err != nil {
		return errors.Wrap(err, "failed to zadd while pushing")
	}
	return nil
}

// Pop pops the message out of the queue. It uses BRPOP underneath, so effectively it blocks until a
// message is available or a timeout is reached.
func (r RedisDriver) Pop(ctx context.Context) (*SerializedMessage, error) {
	if err := r.move(ctx, r.channelConfig.Delayed, r.channelConfig.Waiting); err != nil {
		return nil, err
	}
	if err := r.move(ctx, r.channelConfig.Reserved, r.channelConfig.Timeout); err != nil {
		return nil, err
	}

	res, err := r.redisClient.BRPop(ctx, r.popTimeout, r.channelConfig.Waiting).Result()
	if err == redis.Nil {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to brpop while popping")
	}
	data := res[1]
	var message SerializedMessage
	err = r.packer.Decompress([]byte(data), &message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress message")
	}
	_, err = r.redisClient.ZAdd(ctx, r.channelConfig.Reserved, &redis.Z{
		Score:  float64(time.Now().Add(message.HandleTimeout).Unix()),
		Member: data,
	}).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to zadd while putting message on the reserved queue")
	}
	return &message, nil

}

// Ack acknowledges a message has been processed.
func (r RedisDriver) Ack(ctx context.Context, message *SerializedMessage) error {
	data, err := r.packer.Compress(message)
	if err != nil {
		return errors.Wrap(err, "failed to compress message")
	}
	return r.remove(ctx, r.channelConfig.Reserved, data)
}

// Fail marks a message has failed.
func (r RedisDriver) Fail(ctx context.Context, message *SerializedMessage) error {
	p := r.redisClient.TxPipeline()
	data, err := r.packer.Compress(message)
	if err != nil {
		return errors.Wrap(err, "failed to compress message")
	}
	p.ZRem(ctx, r.channelConfig.Reserved, data)
	message.Attempts++
	data, err = r.packer.Compress(message)
	if err != nil {
		return errors.Wrap(err, "failed to compress message")
	}
	p.LPush(ctx, r.channelConfig.Failed, data)
	_, err = p.Exec(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to lpush while failing message")
	}
	return nil
}

// Reload put failed/timeout message back to the Waiting queue. If the temporary outage have been cleared,
// messages can be tried again via Reload. Reload is not a normal retry.
// It similarly gives otherwise dead messages one more chance,
// but this chance is not subject to the limit of MaxAttempts, nor does it reset the number of time attempted.
func (r RedisDriver) Reload(ctx context.Context, channel string) (int64, error) {
	if channel != r.channelConfig.Failed && channel != r.channelConfig.Timeout {
		return 0, fmt.Errorf("reloading %s is not allowed", channel)
	}
	var count int64 = 0
	for {
		_, err := r.redisClient.RPopLPush(ctx, channel, r.channelConfig.Waiting).Result()
		if errors.Is(err, redis.Nil) {
			break
		}
		if err != nil {
			return count, errors.Wrapf(err, "failed to rpoplpush %s while reloading", channel)
		}
		count++
	}
	return count, nil
}

// Flush flushes a queue of choice by deleting all its data. Use with caution.
func (r RedisDriver) Flush(ctx context.Context, channel string) error {
	_, err := r.redisClient.Del(ctx, channel).Result()
	if err != nil {
		return errors.Wrapf(err, "failed to flush %s", channel)
	}
	return nil
}

type attempt struct {
	err error
}

func (a attempt) try(cmd *redis.IntCmd, value *int64) {
	if a.err != nil && !errors.Is(a.err, redis.Nil) {
		return
	}
	*value, a.err = cmd.Result()
}

// Info lists QueueInfo by inspecting queues one by one. Useful for metrics and monitor.
func (r RedisDriver) Info(ctx context.Context) (QueueInfo, error) {
	var (
		oneByOne attempt
		info     QueueInfo
	)
	oneByOne.try(r.redisClient.LLen(ctx, r.channelConfig.Waiting), &info.Waiting)
	oneByOne.try(r.redisClient.LLen(ctx, r.channelConfig.Failed), &info.Failed)
	oneByOne.try(r.redisClient.LLen(ctx, r.channelConfig.Timeout), &info.Timeout)
	oneByOne.try(r.redisClient.ZCard(ctx, r.channelConfig.Delayed), &info.Delayed)

	if oneByOne.err != nil {
		return info, errors.Wrap(oneByOne.err, "failed to collect queue info")
	}
	return info, nil
}

func (r RedisDriver) remove(ctx context.Context, channel string, data []byte) error {
	_, err := r.redisClient.ZRem(ctx, channel, string(data)).Result()
	if err != nil {
		return errors.Wrapf(err, "failed to zrem while removing from %s", channel)
	}
	return nil
}

// Retry put the message back onto the delayed queue. The message will be tried after a period of time specified
// by Backoff. Note: if one listener failed, all listeners for this event will have to be retried. Make sure
// your listeners are idempotent as always.
func (r RedisDriver) Retry(ctx context.Context, message *SerializedMessage) error {
	p := r.redisClient.TxPipeline()
	data, err := r.packer.Compress(message)
	if err != nil {
		return errors.Wrap(err, "failed to compress message")
	}
	p.ZRem(ctx, r.channelConfig.Reserved, string(data))
	message.Backoff = getRetryDuration(message.Backoff)
	message.Attempts++
	delay := time.Now().Add(message.Backoff)
	data, err = r.packer.Compress(message)
	if err != nil {
		return errors.Wrap(err, "failed to compress message")
	}
	p.ZAdd(ctx, r.channelConfig.Delayed, &redis.Z{
		Score:  float64(delay.Unix()),
		Member: data,
	})
	_, err = p.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to add zset while retrying")
	}
	return nil
}

func (r RedisDriver) move(ctx context.Context, fromKey string, toKey string) error {
	jobs, _ := r.redisClient.ZRevRangeByScore(ctx, fromKey, &redis.ZRangeBy{
		Min:    "-INF",
		Max:    fmt.Sprintf("%d", time.Now().Unix()),
		Offset: 0,
		Count:  100,
	}).Result()
	p := r.redisClient.TxPipeline()
	for _, job := range jobs {
		p.ZRem(ctx, fromKey, job)
		p.LPush(ctx, toKey, job)
	}
	_, err := p.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "move failed")
	}
	return nil
}

func getRetryDuration(d time.Duration) time.Duration {
	d *= 2
	jitter := rand.Float64() + 0.5
	d = time.Duration(int64(float64(d.Nanoseconds()) * jitter))
	if d > 10*time.Minute {
		d = time.Minute
	}
	if d < time.Second {
		d = time.Second
	}
	return d
}

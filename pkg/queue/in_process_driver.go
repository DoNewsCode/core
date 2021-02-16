package queue

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

// An item is something we manage in a priority queue.
type item struct {
	event    *PersistedEvent // The value of the item; arbitrary.
	priority time.Time       // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A priorityQueue implements heap.Interface and holds *PersistedEvent.
type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority.Before(pq[j].priority)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// InProcessDriver is a test replacement for redis driver. It doesn't persist your event in any way,
// so not suitable for production use.
type InProcessDriver struct {
	popInterval time.Duration
	mutex       sync.Mutex
	delayed     *priorityQueue
	waiting     chan *PersistedEvent
	reserved    map[*PersistedEvent]time.Time
	failed      map[*PersistedEvent]struct{}
	timeout     map[*PersistedEvent]struct{}
}

// NewInProcessDriverWithPopInterval creates an *InProcessDriver for testing
func NewInProcessDriver() *InProcessDriver {
	delayed := make(priorityQueue, 0, 10)
	return &InProcessDriver{
		popInterval: time.Second,
		delayed:     &delayed,
		reserved:    make(map[*PersistedEvent]time.Time),
		waiting:     make(chan *PersistedEvent, 1000),
		failed:      make(map[*PersistedEvent]struct{}),
		timeout:     make(map[*PersistedEvent]struct{}),
	}
}

// NewInProcessDriverWithPopInterval creates an *InProcessDriver with an pop interval.
func NewInProcessDriverWithPopInterval(duration time.Duration) *InProcessDriver {
	delayed := make(priorityQueue, 0, 10)
	return &InProcessDriver{
		popInterval: duration,
		delayed:     &delayed,
		reserved:    make(map[*PersistedEvent]time.Time),
		waiting:     make(chan *PersistedEvent, 1000),
		failed:      make(map[*PersistedEvent]struct{}),
		timeout:     make(map[*PersistedEvent]struct{}),
	}
}

func (i *InProcessDriver) Push(ctx context.Context, message *PersistedEvent, delay time.Duration) error {
	if delay > 0 {
		i.mutex.Lock()
		heap.Push(i.delayed, &item{
			event:    message,
			priority: time.Now().Add(delay),
		})
		i.mutex.Unlock()
		return nil
	}
	select {
	case i.waiting <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (i *InProcessDriver) Pop(ctx context.Context) (*PersistedEvent, error) {
	i.mutex.Lock()
	for {
		if len(*i.delayed) == 0 {
			break
		}
		top := heap.Pop(i.delayed).(*item)
		if top.priority.After(time.Now()) {
			heap.Push(i.delayed, top)
			break
		}
		i.waiting <- top.event
	}
	for k := range i.reserved {
		if i.reserved[k].Before(time.Now()) {
			i.timeout[k] = struct{}{}
			delete(i.reserved, k)
		}
	}
	i.mutex.Unlock()
	select {
	case message := <-i.waiting:
		i.reserved[message] = time.Now().Add(message.HandleTimeout)
		return message, nil
	case <-time.After(i.popInterval):
		return nil, Nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (i *InProcessDriver) Ack(ctx context.Context, message *PersistedEvent) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	delete(i.reserved, message)
	return nil
}

func (i *InProcessDriver) Fail(ctx context.Context, message *PersistedEvent) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	delete(i.reserved, message)
	i.failed[message] = struct{}{}
	return nil
}

func (i *InProcessDriver) Reload(ctx context.Context, channel string) (int64, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	var j int64
	if channel == "failed" {
		for k := range i.failed {
			delete(i.failed, k)
			i.waiting <- k
			j++
		}
		return j, nil
	}
	if channel == "timeout" {
		for k := range i.timeout {
			delete(i.timeout, k)
			i.waiting <- k
			j++
		}
		return j, nil
	}
	return 0, fmt.Errorf("unsupported channel %s", channel)
}

func (i *InProcessDriver) Flush(ctx context.Context, channel string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if channel == "failed" {
		i.failed = make(map[*PersistedEvent]struct{})
	}
	if channel == "timeout" {
		i.timeout = make(map[*PersistedEvent]struct{})
	}
	return nil
}

func (i *InProcessDriver) Info(ctx context.Context) (QueueInfo, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return QueueInfo{
		Waiting: int64(len(i.waiting)),
		Delayed: int64(len(*i.delayed)),
		Timeout: int64(len(i.timeout)),
		Failed:  int64(len(i.failed)),
	}, nil
}

func (i *InProcessDriver) Retry(ctx context.Context, message *PersistedEvent) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	delete(i.reserved, message)
	newBackOff := getRetryDuration(message.Backoff)
	heap.Push(i.delayed, &item{
		event:    message,
		priority: time.Now().Add(newBackOff),
	})
	return nil
}

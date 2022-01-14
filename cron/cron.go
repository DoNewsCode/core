package cron

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Cron struct {
	parser            cron.ScheduleParser
	lock              *sync.Cond
	jobDescriptors    jobDescriptors
	globalMiddleware  []JobMiddleware
	location          *time.Location
	nextID            int
	quitWaiter        sync.WaitGroup
	baseContext       context.Context
	baseContextCancel func()
}

func New(options ...Option) *Cron {
	c := &Cron{
		parser:      cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		location:    time.Local,
		nextID:      1,
		lock:        sync.NewCond(&sync.Mutex{}),
		baseContext: context.Background(),
	}
	for _, f := range options {
		f(c)
	}
	return c
}

func (c *Cron) Add(spec string, runner func(ctx context.Context) error, middleware ...JobMiddleware) (JobID, error) {
	schedule, err := c.parser.Parse(spec)
	if err != nil {
		return 0, err
	}

	jobDescriptor := JobDescriptor{
		RawSpec:  spec,
		Run:      runner,
		Schedule: schedule,
		next:     schedule.Next(c.now()),
	}

	middleware = append(middleware, c.globalMiddleware...)

	for i := len(middleware) - 1; i > 0; i-- {
		jobDescriptor = middleware[i](jobDescriptor)
	}

	c.lock.L.Lock()
	defer c.lock.L.Unlock()
	defer c.lock.Broadcast()

	jobDescriptor.ID = JobID(c.nextID)
	c.nextID++

	if jobDescriptor.Name == "" {
		jobDescriptor.Name = fmt.Sprintf("job-%d", jobDescriptor.ID)
	}

	heap.Push(&c.jobDescriptors, &jobDescriptor)
	return jobDescriptor.ID, nil
}

func (c *Cron) Remove(id JobID) {
	c.lock.L.Lock()
	defer c.lock.L.Unlock()

	for i, descriptor := range c.jobDescriptors {
		if descriptor.ID == id {
			heap.Remove(&c.jobDescriptors, i)
		}
	}
}

func (c *Cron) Run(ctx context.Context) error {
	defer c.quitWaiter.Wait()

	c.lock.L.Lock()
	now := c.now()
	for _, descriptor := range c.jobDescriptors {
		descriptor.next = descriptor.Schedule.Next(now)
	}
	heap.Init(&c.jobDescriptors)
	c.lock.L.Unlock()

	var once sync.Once
	for {
		c.lock.L.Lock()
		// Determine the next entry to run.
		for c.jobDescriptors.Len() == 0 || c.jobDescriptors[0].next.IsZero() {
			c.broadcastAtDeadlineOnce(ctx, &once)
			select {
			case <-ctx.Done():
				c.lock.L.Unlock()
				return ctx.Err()
			default:
				c.lock.Wait()
			}
		}
		gap := c.jobDescriptors[0].next.Sub(now)
		c.lock.L.Unlock()

		timer := time.NewTimer(gap)

		select {
		case now = <-timer.C:
			c.lock.L.Lock()
			for {
				if c.jobDescriptors[0].next.After(now) || c.jobDescriptors[0].next.IsZero() {
					break
				}
				descriptor := heap.Pop(&c.jobDescriptors)

				descriptor.(*JobDescriptor).prev = descriptor.(*JobDescriptor).next
				descriptor.(*JobDescriptor).next = descriptor.(*JobDescriptor).Schedule.Next(now)
				heap.Push(&c.jobDescriptors, descriptor)

				var innerCtx context.Context
				innerCtx = context.WithValue(ctx, prevContextKey, descriptor.(*JobDescriptor).prev)
				innerCtx = context.WithValue(innerCtx, nextContextKey, descriptor.(*JobDescriptor).next)

				c.quitWaiter.Add(1)
				go func() {
					defer c.quitWaiter.Done()
					descriptor.(*JobDescriptor).Run(innerCtx)
				}()
			}
			c.lock.L.Unlock()
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (c *Cron) now() time.Time {
	return time.Now().In(c.location)
}

func (c *Cron) broadcastAtDeadlineOnce(ctx context.Context, once *sync.Once) {
	once.Do(func() {
		go func() {
			<-ctx.Done()
			c.lock.Broadcast()
		}()
	})
}

type jobDescriptors []*JobDescriptor

func (j *jobDescriptors) Len() int {
	return len(*j)
}

func (j *jobDescriptors) Less(i, k int) bool {
	return (*j)[i].next.Before((*j)[k].next)
}

func (j *jobDescriptors) Swap(i, k int) {
	(*j)[i], (*j)[k] = (*j)[k], (*j)[i]
}

func (j *jobDescriptors) Push(x interface{}) {
	*j = append(*j, x.(*JobDescriptor))
}

func (j *jobDescriptors) Pop() (v interface{}) {
	*j, v = (*j)[:j.Len()-1], (*j)[j.Len()-1]
	return v
}

type JobID int

type JobDescriptor struct {
	ID       JobID
	Name     string
	RawSpec  string
	Schedule cron.Schedule
	next     time.Time
	prev     time.Time
	Run      func(ctx context.Context) error
}

// Package cron is a partial rewrite based on the github.com/robfig/cron/v3
// package. The API in this package enforces context passing and error
// propagation, and consequently enables better logging, metrics and tracing
// support.
package cron

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Cron schedules jobs to be run on the specified schedule.
type Cron struct {
	parser           cron.ScheduleParser
	lock             *sync.Cond
	jobDescriptors   jobDescriptors
	globalMiddleware []JobOptions
	location         *time.Location
	nextID           int
	quitWaiter       sync.WaitGroup
}

// New returns a new Cron instance.
func New(config Config) *Cron {
	c := &Cron{
		parser:           config.Parser,
		lock:             sync.NewCond(&sync.Mutex{}),
		jobDescriptors:   jobDescriptors{},
		globalMiddleware: config.GlobalOptions,
		location:         config.Location,
		nextID:           1,
	}
	if config.Parser == nil {
		if config.EnableSeconds {
			c.parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		} else {
			c.parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		}
	}
	if config.Location == nil {
		c.location = time.Local
	}
	return c
}

// Add adds a new job to the cron scheduler.A list of middleware can be supplied.
// Note the error returned by the runner will be discarded. It is the user's
// responsibility to handle the error via middleware.
func (c *Cron) Add(spec string, runner func(ctx context.Context) error, middleware ...JobOptions) (JobID, error) {
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

	middleware = append(c.globalMiddleware, middleware...)

	for i := len(middleware) - 1; i >= 0; i-- {
		middleware[i](&jobDescriptor)
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

// Remove removes a job from the cron scheduler.
func (c *Cron) Remove(id JobID) {
	c.lock.L.Lock()
	defer c.lock.L.Unlock()

	for i, descriptor := range c.jobDescriptors {
		if descriptor.ID == id {
			heap.Remove(&c.jobDescriptors, i)
		}
	}
}

// Descriptors returns a list of all job descriptors.
func (c *Cron) Descriptors() []JobDescriptor {
	var descriptors []JobDescriptor

	c.lock.L.Lock()
	defer c.lock.L.Unlock()

	for _, descriptor := range c.jobDescriptors {
		descriptors = append(descriptors, *descriptor)
	}
	return descriptors
}

// Run starts the cron scheduler. It is a blocking call.
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

// JobID is the identifier of jobs.
type JobID int

// JobDescriptor contains the information about jobs.
type JobDescriptor struct {
	// ID is the identifier of job
	ID JobID
	// Name is an optional field typically added by WithName. It can be useful in logging and metrics.
	Name string
	// RawSpec contains the string format of cron schedule format.
	RawSpec string
	// Schedule is the parsed version of RawSpec. It can be overridden by WithSchedule.
	Schedule cron.Schedule
	// Run is the actual work to be done.
	Run func(ctx context.Context) error
	// next is the next time the job should run.
	next time.Time
	// prev is the last time the job ran.
	prev time.Time
}

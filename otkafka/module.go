package otkafka

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/log"
	"github.com/oklog/run"
)

const defaultInterval = 15 * time.Second

// Module is the registration unit for package core.
type Module struct {
	readerMaker     ReaderMaker
	writerMaker     WriterMaker
	env             contract.Env
	logger          log.Logger
	container       contract.Container
	readerCollector *readerCollector
	writerCollector *writerCollector
	interval        time.Duration
	dispatcher      contract.Dispatcher
}

type moduleIn struct {
	di.In

	ReaderMaker     ReaderMaker
	WriterMaker     WriterMaker
	Env             contract.Env
	Logger          log.Logger
	Container       contract.Container
	ReaderCollector *readerCollector
	WriterCollector *writerCollector
	Conf            contract.ConfigAccessor
	Dispatcher      contract.Dispatcher `optional:"true"`
}

// New creates a Module.
func New(in moduleIn) Module {
	var duration time.Duration = defaultInterval
	in.Conf.Unmarshal("kafkaMetrics.interval", &duration)
	return Module{
		readerMaker:     in.ReaderMaker,
		writerMaker:     in.WriterMaker,
		env:             in.Env,
		logger:          in.Logger,
		container:       in.Container,
		readerCollector: in.ReaderCollector,
		writerCollector: in.WriterCollector,
		interval:        duration,
		dispatcher:      in.Dispatcher,
	}
}

// ProvideRunGroup add a goroutine to periodically scan kafka's reader&writer info and
// report them to metrics collector such as prometheus.
func (m Module) ProvideRunGroup(group *run.Group) {
	if m.readerCollector != nil || m.writerCollector != nil {
		ctx, cancel := context.WithCancel(context.Background())
		ticker := time.NewTicker(m.interval)
		group.Add(func() error {
			for {
				select {
				case <-ticker.C:
					if m.readerCollector != nil {
						m.readerCollector.collectConnectionStats()
					}
					if m.writerCollector != nil {
						m.writerCollector.collectConnectionStats()
					}
				case <-ctx.Done():
					ticker.Stop()
					return nil
				}
			}
		}, func(err error) {
			cancel()
		})
		return
	}

	if m.canHotReloadWriter() {
		ctx, cancel := context.WithCancel(context.Background())
		group.Add(func() error {
			return m.writerMaker.(WriterFactory).Watch(ctx, m.dispatcher)
		}, func(err error) {
			cancel()
		})
	}
	if m.canHotReloadReader() {
		ctx, cancel := context.WithCancel(context.Background())
		group.Add(func() error {
			return m.readerMaker.(ReaderFactory).Watch(ctx, m.dispatcher)
		}, func(err error) {
			cancel()
		})
	}
}

func (m Module) canHotReloadReader() bool {
	if m.dispatcher == nil {
		return false
	}
	if _, ok := m.readerMaker.(ReaderFactory); !ok {
		return false
	}
	return true
}

func (m Module) canHotReloadWriter() bool {
	if m.dispatcher == nil {
		return false
	}
	if _, ok := m.writerMaker.(WriterFactory); !ok {
		return false
	}
	return true
}

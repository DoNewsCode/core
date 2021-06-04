package otkafka

import (
	"context"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/log"
	"github.com/oklog/run"
	"time"
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
	}
}

// ProvideRunGroup add a goroutine to periodically scan kafka's reader&writer info and
// report them to metrics collector such as prometheus.
func (m Module) ProvideRunGroup(group *run.Group) {
	if m.readerCollector == nil && m.writerCollector == nil {
		return
	}
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
}

package pool

import (
	"context"
	"sync"

	"github.com/DoNewsCode/core/di"
	"github.com/oklog/run"
)

// Maker models Factory
type Maker interface {
	Make(name string) (*Pool, error)
}

type Factory struct {
	wg *sync.WaitGroup

	factory *di.Factory[*Pool]
}

func (f *Factory) Make(name string) (*Pool, error) {
	return f.factory.Make(name)
}

// ProvideRunGroup implements core.RunProvider
func (f *Factory) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())

	group.Add(func() error {
		f.run(ctx)
		return nil
	}, func(err error) {
		cancel()
	})
}

// Module implements di.Modular
func (f *Factory) Module() interface{} {
	return f
}

func (f *Factory) run(ctx context.Context) {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		<-ctx.Done()
	}()
	f.wg.Wait()
}

package sagas

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	var g run.Group
	c := core.Default()
	module := New(In{
		Conf:   config.MapAdapter{"sagas.recoverIntervalSecond": 1.0, "sagas.defaultSagaTimeoutSecond": 1.0},
		Logger: log.NewNopLogger(),
		Store:  &InProcessStore{},
		Sagas: []*Saga{{
			Name: "foo",
			Steps: []*Step{{
				Name: "bar",
				Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
					return nil, nil
				},
				Undo: func(ctx context.Context, req interface{}) error {
					return nil
				},
			}},
		}},
	})
	c.AddModule(module)
	c.ApplyRunGroup(&g)
	timeout(time.Second, &g)
	assert.NoError(t, g.Run())
}

func timeout(duration time.Duration, g *run.Group) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	g.Add(func() error {
		<-ctx.Done()
		return nil
	}, func(err error) {
		cancel()
	})
}

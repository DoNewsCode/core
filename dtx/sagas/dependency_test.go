package sagas

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/oklog/run"
	"github.com/stretchr/testify/assert"
)

type sagas struct {
	di.Out

	Step *Step `group:"saga"`
}

func TestNew(t *testing.T) {
	t.Parallel()
	var g run.Group
	c := core.Default()
	c.Provide(Providers())
	c.Provide(di.Deps{func() sagas {
		return sagas{
			Step: &Step{
				Name: "bar",
				Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
					return 1, nil
				},
				Undo: func(ctx context.Context, req interface{}) (err error) {
					return nil
				},
			},
		}
	}})
	c.Invoke(func(r *Registry, endpoints SagaEndpoints) {
		tx, ctx := r.StartTX(context.Background())
		resp, _ := endpoints["bar"](ctx, nil)
		assert.Equal(t, 1, resp)
		tx.Commit(ctx)
		c.ApplyRunGroup(&g)
		timeout(time.Second, &g)
		assert.NoError(t, g.Run())
	})
}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	assert.NotNil(t, conf)
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

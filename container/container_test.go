package container

import (
	"testing"

	"github.com/DoNewsCode/core/cron"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mock struct{}

func (m mock) ProvideRunGroup(group *run.Group) {
	panic("implement me")
}

func (m mock) ProvideGRPC(server *grpc.Server) {
	panic("implement me")
}

func (m mock) ProvideHTTP(router *mux.Router) {
	panic("implement me")
}

func (m mock) ProvideCron(crontab *cron.Cron) {
	panic("implement me")
}

func (m mock) ProvideCommand(command *cobra.Command) {
	panic("implement me")
}

func TestContainer_AddModule(t *testing.T) {
	cases := []struct {
		name    string
		module  interface{}
		asserts func(t *testing.T, container Container)
	}{
		{
			"any",
			"foo",
			func(t *testing.T, container Container) {
				assert.Contains(t, container.modules, "foo")
			},
		},
		{
			"mock",
			mock{},
			func(t *testing.T, container Container) {
				assert.Len(t, container.runProviders, 1)
				assert.Len(t, container.httpProviders, 1)
				assert.Len(t, container.grpcProviders, 1)
				assert.Len(t, container.cronProviders, 1)
				assert.Len(t, container.commandProviders, 1)
				assert.Len(t, container.closerProviders, 0)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var container Container
			container.AddModule(c.module)
			c.asserts(t, container)
		})
	}
}

type closer func()

func (f closer) ProvideCloser() {
	f()
}

func TestContainer_Shutdown(t *testing.T) {
	seq := 0
	container := Container{}
	container.AddModule(closer(func() { assert.Equal(t, 2, seq); seq = 1 }))
	container.AddModule(closer(func() { assert.Equal(t, 3, seq); seq = 2 }))
	container.AddModule(closer(func() { assert.Equal(t, 0, seq); seq = 3 }))
	container.Shutdown()
	assert.Equal(t, 1, seq)
}

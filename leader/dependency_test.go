package leader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/DoNewsCode/core/otetcd"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/dig"
)

type mockMaker struct {
	name      string
	endpoints []string
}

func (m *mockMaker) Make(name string) (*clientv3.Client, error) {
	m.name = name
	return clientv3.New(clientv3.Config{Endpoints: m.endpoints})
}

type mockDriver struct{}

func (m mockDriver) Campaign(ctx context.Context, toLeader func(bool)) error {
	panic("implement me")
}

func (m mockDriver) Resign(ctx context.Context) error {
	panic("implement me")
}

func TestDriverConstructorsAndDriverPriority(t *testing.T) {
	driver := mockDriver{}
	ctor := func(args DriverArgs) (Driver, error) {
		return mockDriver{}, nil
	}
	out, _ := provide(&providersOption{
		driver:            driver,
		driverConstructor: ctor,
	})(in{})
	assert.Equal(t, driver, out.Election.driver)
}

func TestDriverConstructor(t *testing.T) {
	ctor := func(args DriverArgs) (Driver, error) {
		return mockDriver{}, nil
	}
	out, _ := provide(&providersOption{
		driver:            nil,
		driverConstructor: ctor,
	})(in{})
	assert.IsType(t, mockDriver{}, out.Election.driver)
}

func TestFailedDriverConstructor(t *testing.T) {
	ctor := func(args DriverArgs) (Driver, error) {
		return nil, fmt.Errorf("failed")
	}
	_, err := provide(&providersOption{
		driver:            nil,
		driverConstructor: ctor,
	})(in{})
	assert.Error(t, err)
}

type mockPopulator struct {
	endpoints []string
}

func (m mockPopulator) Populate(target any) error {
	c := dig.New()
	c.Provide(func() contract.AppName {
		return config.AppName("foo")
	})
	c.Provide(func() contract.Env {
		return config.EnvUnknown
	})
	c.Provide(func() otetcd.Maker {
		return &mockMaker{"default", m.endpoints}
	})
	c.Provide(func() contract.ConfigUnmarshaler {
		return config.MapAdapter{}
	})
	p := di.IntoPopulator(c)
	return p.Populate(target)
}

func TestDefaultDriver(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("Set env ETCD_ADDR to run TestDefaultDriver")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	out, err := provide(
		&providersOption{
			driver:            nil,
			driverConstructor: nil,
		},
	)(
		in{
			Config:    config.MapAdapter{"etcdName": "default"},
			Populator: mockPopulator{addrs},
		},
	)
	assert.NoError(t, err)
	assert.IsType(t, &leaderetcd.EtcdDriver{}, out.Election.driver)
}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	assert.NotNil(t, conf)
}

func TestPreferDriverInDI(t *testing.T) {
	g := dig.New()
	g.Provide(func() Driver {
		return mockDriver{}
	})
	driver, err := newDefaultDriver(DriverArgs{
		Populator: di.IntoPopulator(g),
	})
	assert.NoError(t, err)
	assert.IsType(t, mockDriver{}, driver)
}

func TestPreferDriverInDI_error(t *testing.T) {
	g := dig.New()
	g.Provide(func() (Driver, error) {
		return mockDriver{}, errors.New("err")
	})
	_, err := newDefaultDriver(DriverArgs{
		Populator: di.IntoPopulator(g),
	})
	assert.Error(t, err)
}

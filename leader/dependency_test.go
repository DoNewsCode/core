package leader

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/DoNewsCode/core/otetcd"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

type mockMaker struct {
	name      string
	endpoints []string
}

func (m *mockMaker) Make(name string) (*clientv3.Client, error) {
	m.name = name
	return clientv3.New(clientv3.Config{Endpoints: m.endpoints})
}

func TestDrivers(t *testing.T) {
	driver := mockDriver{}
	out, _ := provide(&providersOption{
		driver:            nil,
		driverConstructor: nil,
	})(in{})
	assert.Equal(t, driver, out.Election.driver)
}

type mockDriver struct{}

func (m mockDriver) Campaign(ctx context.Context) error {
	panic("implement me")
}

func (m mockDriver) Resign(ctx context.Context) error {
	panic("implement me")
}

func TestDriverConstructorsAndDriverPriority(t *testing.T) {
	driver := mockDriver{}
	ctor := func(args DriverConstructorArgs) (Driver, error) {
		return mockDriver{}, nil
	}
	out, _ := provide(&providersOption{
		driver:            driver,
		driverConstructor: ctor,
	})(in{})
	assert.Equal(t, driver, out.Election.driver)
}

func TestDriverConstructor(t *testing.T) {
	ctor := func(args DriverConstructorArgs) (Driver, error) {
		return mockDriver{}, nil
	}
	out, _ := provide(&providersOption{
		driver:            nil,
		driverConstructor: ctor,
	})(in{})
	assert.IsType(t, mockDriver{}, out.Election.driver)
}

type mockPopulater struct {
	endpoints []string
}

func (m mockPopulater) Populate(target interface{}) error {
	*target.(*otetcd.Maker) = &mockMaker{"default", m.endpoints}
	return nil
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
			Populater: mockPopulater{addrs},
			AppName:   config.AppName("test"),
			Env:       config.NewEnv("test"),
		},
	)
	assert.NoError(t, err)
	assert.IsType(t, &leaderetcd.EtcdDriver{}, out.Election.driver)
}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	assert.NotNil(t, conf)
}

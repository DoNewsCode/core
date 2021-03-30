package leader

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	leaderetcd2 "github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

type mockMaker struct {
	name string
}

func (m *mockMaker) Make(name string) (*clientv3.Client, error) {
	m.name = name
	return clientv3.New(clientv3.Config{})
}

func TestDetermineDriver(t *testing.T) {
	driver := &leaderetcd2.EtcdDriver{}
	p := in{}
	p.Driver = driver
	determineDriver(&p)
	assert.Equal(t, driver, p.Driver)

	maker := &mockMaker{}
	p = in{
		Config: config.MapAdapter{
			"leader": Option{
				EtcdName: "",
			},
		},
		Dispatcher: nil,
		Driver:     nil,
		Maker:      maker,
	}
	determineDriver(&p)
	assert.Equal(t, "default", maker.name)

	p = in{
		Config: config.MapAdapter{
			"leader": Option{
				EtcdName: "foo",
			},
		},
		Dispatcher: nil,
		Driver:     nil,
		Maker:      maker,
	}
	determineDriver(&p)
	assert.Equal(t, "foo", maker.name)

	p = in{
		Config: config.MapAdapter{
			"leader": Option{
				EtcdName: "foo",
			},
		},
		Dispatcher: nil,
		Driver:     nil,
		Maker:      nil,
		AppName:    config.AppName("foo"),
		Env:        config.EnvTesting,
	}
	err := determineDriver(&p)
	assert.Error(t, err)

	p = in{
		Config:     config.MapAdapter{},
		Dispatcher: nil,
		Driver:     nil,
		Maker:      maker,
		AppName:    config.AppName("foo"),
		Env:        config.EnvTesting,
	}
	err = determineDriver(&p)
	assert.Error(t, err)
}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	_, err := yaml.Marshal(conf.Config)
	assert.NoError(t, err)
}

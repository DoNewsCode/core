package leader

import (
	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
	"testing"
)

type mockMaker struct {
	name string
}

func (m *mockMaker) Make(name string) (*clientv3.Client, error) {
	m.name = name
	return clientv3.New(clientv3.Config{})
}

func TestDetermineDriver(t *testing.T) {
	driver := &EtcdDriver{}
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
		Env:        config.Env("testing"),
	}
	err := determineDriver(&p)
	assert.Error(t, err)

	p = in{
		Config:     config.MapAdapter{},
		Dispatcher: nil,
		Driver:     nil,
		Maker:      maker,
		AppName:    config.AppName("foo"),
		Env:        config.Env("testing"),
	}
	err = determineDriver(&p)
	assert.Error(t, err)

}

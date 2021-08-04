package leader

import (
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

type mockMaker struct {
	name string
}

func (m *mockMaker) Make(name string) (*clientv3.Client, error) {
	m.name = name
	return clientv3.New(clientv3.Config{})
}

func TestDetermineDriver(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestDetermineDriver")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	client, _ := clientv3.New(clientv3.Config{
		Endpoints: addrs,
	})
	driver := leaderetcd.NewEtcdDriver(client, key.New())
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
	assert.NotNil(t, conf)
}

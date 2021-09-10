package leader

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/DoNewsCode/core/otetcd"
	"github.com/oklog/run"
)

/*
Providers returns a set of dependency providers for *Election and *Status.
	Depends On:
		contract.ConfigAccessor
		contract.Dispatcher
		contract.DIPopulator
	Provide:
		Election *Election
		Status   *Status
*/
func Providers(opt ...ProvidersOptionFunc) di.Deps {
	option := &providersOption{
		driver:            nil,
		driverConstructor: nil,
	}
	for _, f := range opt {
		f(option)
	}
	return di.Deps{provide(option), provideConfig}
}

type in struct {
	di.In

	Config     contract.ConfigUnmarshaler
	Dispatcher contract.Dispatcher
	Populator  contract.DIPopulator
}

type out struct {
	di.Out

	Election *Election
	Status   *Status
}

func provide(option *providersOption) func(in in) (out, error) {
	return func(in in) (out, error) {
		if option.driver != nil {
			e := NewElection(in.Dispatcher, option.driver)
			return out{
				Election: e,
				Status:   e.status,
			}, nil
		}

		driverConstructor := newDefaultDriver
		if option.driverConstructor != nil {
			driverConstructor = option.driverConstructor
		}
		driver, err := driverConstructor(DriverArgs{
			Populator: in.Populator,
		})
		if err != nil {
			return out{}, fmt.Errorf("unable to contruct driver: %w", err)
		}

		e := NewElection(in.Dispatcher, driver)
		return out{Election: e, Status: e.status}, nil
	}

}

// Module marks out as a module.
func (m out) Module() interface{} { return m }

func (m out) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(func() error {
		err := m.Election.Campaign(ctx)
		if err != nil {
			return err
		}
		<-ctx.Done()
		return nil
	}, func(err error) {
		_ = m.Election.Resign(ctx)
		cancel()
	})
}

// DriverArgs is the argument for constructing new drivers.
type DriverArgs struct {
	Populator contract.DIPopulator
}

func newDefaultDriver(args DriverArgs) (Driver, error) {
	var injected struct {
		di.In

		Conf    contract.ConfigUnmarshaler
		AppName contract.AppName
		Env     contract.Env
		Maker   otetcd.Maker
	}
	if err := args.Populator.Populate(&injected); err != nil {
		return nil, fmt.Errorf("missing dependency for the default election driver: %w", err)
	}
	var option Option
	if err := injected.Conf.Unmarshal("leader", &option); err != nil {
		return nil, fmt.Errorf("leader election configuration error: %w", err)
	}
	if option.EtcdName == "" {
		option.EtcdName = "default"
	}
	etcdClient, err := injected.Maker.Make(option.EtcdName)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate leader election with etcd driver (%s): %w", option.EtcdName, err)
	}
	return leaderetcd.NewEtcdDriver(etcdClient, key.New(injected.AppName.String(), injected.Env.String())), nil
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func provideConfig() configOut {
	return configOut{Config: []config.ExportedConfig{
		{
			Owner: "leader",
			Data: map[string]interface{}{
				"leader": map[string]interface{}{
					"etcdName": "default",
				},
			},
			Comment: "The leader election config",
		},
	}}
}

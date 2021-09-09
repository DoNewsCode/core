package leader

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/key"
	leaderetcd2 "github.com/DoNewsCode/core/leader/leaderetcd"
	"github.com/DoNewsCode/core/otetcd"
	"github.com/oklog/run"
)

/*
Providers returns a set of dependency providers for *Election and *Status.
	Depends On:
		contract.AppName
		contract.Env
		contract.ConfigAccessor
		contract.Dispatcher
		Driver       `optional:"true"`
		otetcd.Maker `optional:"true"`
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
	return []interface{}{provide(option), provideConfig}
}

type in struct {
	di.In

	AppName    contract.AppName
	Env        contract.Env
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

		driver, err := driverConstructor(DriverConstructorArgs{
			Conf:      in.Config,
			AppName:   in.AppName,
			Env:       in.Env,
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

// DriverConstructorArgs is the argument for constructing new drivers.
type DriverConstructorArgs struct {
	Conf      contract.ConfigUnmarshaler
	AppName   contract.AppName
	Env       contract.Env
	Populator contract.DIPopulator
}

func newDefaultDriver(args DriverConstructorArgs) (Driver, error) {
	var option Option
	if err := args.Conf.Unmarshal("leader", &option); err != nil {
		return nil, fmt.Errorf("leader election configuration error: %w", err)
	}
	if option.EtcdName == "" {
		option.EtcdName = "default"
	}
	var maker otetcd.Maker
	args.Populator.Populate(&maker)
	if maker == nil {
		return nil, fmt.Errorf("must provider an otetcd.Maker to the DI graph to construct the default leader.Driver")
	}
	etcdClient, err := maker.Make(option.EtcdName)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate leader election with etcd driver (%s): %w", option.EtcdName, err)
	}
	return leaderetcd2.NewEtcdDriver(etcdClient, key.New(args.AppName.String(), args.Env.String())), nil
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

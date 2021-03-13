package otetcd

import (
	"fmt"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

/*
Providers returns a set of dependencies including the Maker, the default *clientv3.Client and the exported configs.
	Depends On:
		log.Logger
		contract.ConfigAccessor
		EtcdConfigInterceptor `optional:"true"`
		opentracing.Tracer    `optional:"true"`
	Provide:
		Maker
		Factory
		*clientv3.Client
*/
func Providers() []interface{} {
	return []interface{}{provideFactory, provideDefaultClient, provideExportedConfigs}
}

// EtcdConfigInterceptor is an injector type hint that allows user to do
// last minute modification to etcd configurations. This is useful when some
// configuration can not be expressed in yaml/json. For example, the *tls.Config.
type EtcdConfigInterceptor func(name string, options *clientv3.Config)

// Maker is models Factory
type Maker interface {
	Make(name string) (*clientv3.Client, error)
}

// Factory is a *di.Factory that creates redis.UniversalClient using a
// specific configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates redis.UniversalClient using a specific configuration entry.
func (r Factory) Make(name string) (*clientv3.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*clientv3.Client), nil
}

// factoryIn is the injection parameter for provideFactory.
type factoryIn struct {
	di.In

	Logger      log.Logger
	Conf        contract.ConfigAccessor
	Interceptor EtcdConfigInterceptor `optional:"true"`
	Tracer      opentracing.Tracer    `optional:"true"`
}

// FactoryOut is the result of Provide.
type FactoryOut struct {
	di.Out

	Maker   Maker
	Factory Factory
}

// provideFactory creates Factory. It is a valid
// dependency for package core.
func provideFactory(p factoryIn) (FactoryOut, func()) {
	var err error
	var dbConfs map[string]Option

	err = p.Conf.Unmarshal("etcd", &dbConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}

	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			ok   bool
			conf Option
		)
		if conf, ok = dbConfs[name]; !ok {
			if name != "default" {
				return di.Pair{}, fmt.Errorf("etcd configuration %s not valid", name)
			}
			conf = Option{Endpoints: []string{"localhost:2379"}}
		}
		co := clientv3.Config{
			Endpoints:            conf.Endpoints,
			AutoSyncInterval:     duration(conf.AutoSyncIntervalSecond),
			DialTimeout:          duration(conf.DialTimeoutSecond),
			DialKeepAliveTime:    duration(conf.DialKeepAliveTimeSecond),
			DialKeepAliveTimeout: duration(conf.DialKeepAliveTimeoutSecond),
			MaxCallSendMsgSize:   conf.MaxCallSendMsgSize,
			MaxCallRecvMsgSize:   conf.MaxCallRecvMsgSize,
			TLS:                  conf.TLS,
			Username:             conf.Username,
			Password:             conf.Password,
			RejectOldCluster:     conf.RejectOldCluster,
			DialOptions:          conf.DialOptions,
			Context:              conf.Context,
			LogConfig:            conf.LogConfig,
			PermitWithoutStream:  conf.PermitWithoutStream,
		}
		if p.Tracer != nil {
			co.DialOptions = append(
				co.DialOptions,
				grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(p.Tracer)),
				grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(p.Tracer)),
			)
		}
		if p.Interceptor != nil {
			p.Interceptor(name, &co)
		}
		client, _ := clientv3.New(co)
		return di.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Close()
			},
		}, nil
	})
	etcdFactory := Factory{factory}
	out := FactoryOut{
		Maker:   etcdFactory,
		Factory: etcdFactory,
	}
	return out, factory.Close
}

func provideDefaultClient(maker Maker) (*clientv3.Client, error) {
	return maker.Make("default")
}

type configOut struct {
	di.Out

	ExportedConfig []config.ExportedConfig `group:"config,flatten"`
}

func provideExportedConfigs() configOut {
	return configOut{
		ExportedConfig: []config.ExportedConfig{
			{
				"otetcd",
				map[string]interface{}{
					"etcd": map[string]Option{
						"default": {
							Endpoints:                  []string{"127.0.0.1:2379"},
							AutoSyncIntervalSecond:     0,
							DialTimeoutSecond:          0,
							DialKeepAliveTimeSecond:    0,
							DialKeepAliveTimeoutSecond: 0,
							MaxCallSendMsgSize:         0,
							MaxCallRecvMsgSize:         0,
							TLS:                        nil,
							Username:                   "",
							Password:                   "",
							RejectOldCluster:           false,
							DialOptions:                nil,
							Context:                    nil,
							LogConfig:                  nil,
							PermitWithoutStream:        false,
						},
					},
				},
				"The configuration for ETCD.",
			},
		},
	}
}

func duration(sec float64) time.Duration {
	return time.Duration(sec) * time.Second
}

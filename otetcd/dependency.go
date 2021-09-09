package otetcd

import (
	"fmt"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type providersOption struct {
	interceptor EtcdConfigInterceptor
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithConfigInterceptor instructs the Providers to accept the
// EtcdConfigInterceptor so that users can change config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithConfigInterceptor(interceptor EtcdConfigInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.interceptor = interceptor
	}
}

/*
Providers returns a set of dependencies including the Maker, the default *clientv3.Client and the exported configs.
	Depends On:
		log.Logger
		contract.ConfigAccessor
		opentracing.Tracer    `optional:"true"`
	Provide:
		Maker
		Factory
		*clientv3.Client
*/
func Providers(opts ...ProvidersOptionFunc) []interface{} {
	option := providersOption{interceptor: func(name string, options *clientv3.Config) {}}
	for _, f := range opts {
		f(&option)
	}
	return []interface{}{provideFactory(&option), provideDefaultClient, provideConfig}
}

// EtcdConfigInterceptor is an injector type hint that allows user to do
// last minute modification to etcd configurations. This is useful when some
// configuration can not be expressed in yaml/json. For example, the *tls.Config.
type EtcdConfigInterceptor func(name string, options *clientv3.Config)

// factoryIn is the injection parameter for provideFactory.
type factoryIn struct {
	di.In

	Logger     log.Logger
	Conf       contract.ConfigUnmarshaler
	Tracer     opentracing.Tracer  `optional:"true"`
	Dispatcher contract.Dispatcher `optional:"true"`
}

// FactoryOut is the result of Provide.
type FactoryOut struct {
	di.Out

	Maker   Maker
	Factory Factory
}

// provideFactory creates Factory. It is a valid
// dependency for package core.
func provideFactory(option *providersOption) func(p factoryIn) (FactoryOut, func()) {
	if option.interceptor == nil {
		option.interceptor = func(name string, options *clientv3.Config) {}
	}

	return func(p factoryIn) (FactoryOut, func()) {

		factory := di.NewFactory(func(name string) (di.Pair, error) {
			var (
				conf Option
			)
			if err := p.Conf.Unmarshal(fmt.Sprintf("etcd.%s", name), &conf); err != nil {
				return di.Pair{}, fmt.Errorf("etcd configuration %s not valid: %w", name, err)
			}
			if len(conf.Endpoints) == 0 {
				conf.Endpoints = []string{"127.0.0.1:2379"}
			}
			co := clientv3.Config{
				Endpoints:            conf.Endpoints,
				AutoSyncInterval:     duration(conf.AutoSyncInterval),
				DialTimeout:          duration(conf.dialTimeout()),
				DialKeepAliveTime:    duration(conf.DialKeepAliveTime),
				DialKeepAliveTimeout: duration(conf.DialKeepAliveTimeout),
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
			option.interceptor(name, &co)
			client, _ := clientv3.New(co)
			return di.Pair{
				Conn: client,
				Closer: func() {
					_ = client.Close()
				},
			}, nil
		})
		etcdFactory := Factory{factory}
		etcdFactory.SubscribeReloadEventFrom(p.Dispatcher)
		out := FactoryOut{
			Maker:   etcdFactory,
			Factory: etcdFactory,
		}
		return out, factory.Close
	}
}

func provideDefaultClient(maker Maker) (*clientv3.Client, error) {
	return maker.Make("default")
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func provideConfig() configOut {
	return configOut{
		Config: []config.ExportedConfig{
			{
				Owner: "otetcd",
				Data: map[string]interface{}{
					"etcd": map[string]Option{
						"default": {
							Endpoints:            []string{"127.0.0.1:2379"},
							AutoSyncInterval:     config.Duration{},
							DialTimeout:          config.Duration{},
							DialKeepAliveTime:    config.Duration{},
							DialKeepAliveTimeout: config.Duration{},
							MaxCallSendMsgSize:   0,
							MaxCallRecvMsgSize:   0,
							TLS:                  nil,
							Username:             "",
							Password:             "",
							RejectOldCluster:     false,
							DialOptions:          nil,
							Context:              nil,
							LogConfig:            nil,
							PermitWithoutStream:  false,
						},
					},
				},
				Comment: "The configuration for ETCD.",
			},
		},
	}
}

func duration(d config.Duration) time.Duration {
	return d.Duration
}

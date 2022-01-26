package otetcd

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/log"
	"github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type providersOption struct {
	reloadable  bool
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

// WithReload toggles whether the factory should reload cached instances upon
// OnReload event.
func WithReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.reloadable = shouldReload
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
func Providers(opts ...ProvidersOptionFunc) di.Deps {
	option := providersOption{interceptor: func(name string, options *clientv3.Config) {}}
	for _, f := range opts {
		f(&option)
	}
	return di.Deps{
		provideFactory(&option),
		provideDefaultClient,
		provideConfig,
		di.Bind(new(*Factory), new(Maker)),
	}
}

// EtcdConfigInterceptor is an injector type hint that allows user to do
// last minute modification to etcd configurations. This is useful when some
// configuration can not be expressed in yaml/json. For example, the *tls.Config.
type EtcdConfigInterceptor func(name string, options *clientv3.Config)

// factoryIn is the injection parameter for provideFactory.
type factoryIn struct {
	di.In

	Logger        log.Logger
	Conf          contract.ConfigUnmarshaler
	Tracer        opentracing.Tracer              `optional:"true"`
	OnReloadEvent contract.ConfigReloadDispatcher `optional:"true"`
}

// provideFactory creates Factory. It is a valid
// dependency for package core.
func provideFactory(option *providersOption) func(p factoryIn) (*Factory, func()) {
	if option.interceptor == nil {
		option.interceptor = func(name string, options *clientv3.Config) {}
	}

	return func(p factoryIn) (*Factory, func()) {
		factory := di.NewFactory[*clientv3.Client](func(name string) (pair di.Pair[*clientv3.Client], err error) {
			var conf Option
			if err := p.Conf.Unmarshal(fmt.Sprintf("etcd.%s", name), &conf); err != nil {
				return pair, fmt.Errorf("etcd configuration %s not valid: %w", name, err)
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
			return di.Pair[*clientv3.Client]{
				Conn: client,
				Closer: func() {
					_ = client.Close()
				},
			}, nil
		})
		if option.reloadable && p.OnReloadEvent != nil {
			p.OnReloadEvent.Subscribe(func(_ context.Context, _ contract.ConfigUnmarshaler) error {
				factory.Close()
				return nil
			})
		}
		return factory, factory.Close
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

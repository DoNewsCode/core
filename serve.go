package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/cron"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type serveIn struct {
	di.In

	Config             contract.ConfigAccessor
	Logger             log.Logger
	Container          contract.Container
	HTTPServer         *http.Server                 `optional:"true"`
	HTTPRouter         *mux.Router                  `optional:"true"`
	GRPCServer         *grpc.Server                 `optional:"true"`
	HTTPServerStart    lifecycle.HTTPServerStart    `optional:"true"`
	HTTPServerShutdown lifecycle.HTTPServerShutdown `optional:"true"`
	GRPCServerStart    lifecycle.GRPCServerStart    `optional:"true"`
	GRPCServerShutdown lifecycle.GRPCServerShutdown `optional:"true"`
	Cron               *cron.Cron                   `optional:"true"`
}

func NewServeModule(in serveIn) serveModule {
	return serveModule{
		in,
	}
}

var _ CommandProvider = (*serveModule)(nil)

type serveModule struct {
	in serveIn
}

func (s serveModule) ProvideCommand(command *cobra.Command) {
	command.AddCommand(newServeCmd(s.in))
}

type runGroupFunc func(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error)

func (s serveIn) httpServe(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error) {
	type httpConfig struct {
		Disable           bool            `json:"disable" yaml:"disable"`
		Addr              string          `json:"addr" yaml:"addr"`
		ReadTimeout       config.Duration `json:"readTimeout" yaml:"readTimeout"`
		ReadHeaderTimeout config.Duration `json:"readHeaderTimeout" yaml:"readHeaderTimeout"`
		WriteTimeout      config.Duration `json:"writeTimeout" yaml:"writeTimeout"`
		IdleTimeout       config.Duration `json:"idleTimeout" yaml:"idleTimeout"`
		MaxHeaderBytes    int             `json:"maxHeaderBytes" yaml:"maxHeaderBytes"`
	}

	var conf httpConfig
	s.Config.Unmarshal("http", &conf)

	if conf.Disable {
		return nil, nil, nil
	}

	if s.HTTPServer == nil {
		s.HTTPServer = &http.Server{
			ReadTimeout:       conf.ReadTimeout.Duration,
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration,
			WriteTimeout:      conf.WriteTimeout.Duration,
			IdleTimeout:       conf.IdleTimeout.Duration,
			MaxHeaderBytes:    conf.MaxHeaderBytes,
		}
	}
	if s.HTTPRouter == nil {
		s.HTTPRouter = mux.NewRouter()
	}
	applyRouter(s.Container, s.HTTPRouter)

	s.HTTPRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, _ := route.GetPathTemplate()
		level.Debug(logger).Log("tag", "http", "path", tpl)
		return nil
	})

	s.HTTPServer.Handler = s.HTTPRouter
	httpAddr := conf.Addr
	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed start http server")
	}
	return func() error {
			logger.Infof("http service is listening at %s", ln.Addr())
			s.HTTPServerStart.Fire(
				ctx,
				lifecycle.HTTPServerStartPayload{HTTPServer: s.HTTPServer, Listener: ln},
			)
			defer s.HTTPServerShutdown.Fire(
				ctx,
				lifecycle.HTTPServerShutdownPayload{HTTPServer: s.HTTPServer, Listener: ln},
			)
			return s.HTTPServer.Serve(ln)
		}, func(err error) {
			_ = s.HTTPServer.Shutdown(context.Background())
			_ = ln.Close()
		}, nil
}

func (s serveIn) grpcServe(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error) {
	if s.Config.Bool("grpc.disable") {
		return nil, nil, nil
	}
	if s.GRPCServer == nil {
		s.GRPCServer = grpc.NewServer()
	}
	applyGRPCServer(s.Container, s.GRPCServer)

	for module, info := range s.GRPCServer.GetServiceInfo() {
		for _, method := range info.Methods {
			level.Debug(logger).Log("tag", "grpc", "path", fmt.Sprintf("%s/%s", module, method.Name))
		}
	}

	grpcAddr := s.Config.String("grpc.addr")
	ln, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed start grpc server")
	}
	return func() error {
			logger.Infof("gRPC service is listening at %s", ln.Addr())
			s.GRPCServerStart.Fire(
				ctx,
				lifecycle.GRPCServerStartPayload{GRPCServer: s.GRPCServer, Listener: ln},
			)
			defer s.GRPCServerShutdown.Fire(
				ctx,
				lifecycle.GRPCServerShutdownPayload{GRPCServer: s.GRPCServer, Listener: ln},
			)
			return s.GRPCServer.Serve(ln)
		}, func(err error) {
			s.GRPCServer.GracefulStop()
			_ = ln.Close()
		}, nil
}

func (s serveIn) cronServe(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error) {
	if s.Config.Bool("cron.disable") {
		return nil, nil, nil
	}
	if s.Cron == nil {
		s.Cron = cron.New(cron.Config{GlobalOptions: []cron.JobOption{cron.WithLogging(log.With(s.Logger, "tag", "cron"))}})
	}
	applyCron(s.Container, s.Cron)
	if len(s.Cron.Descriptors()) > 0 {
		ctx, cancel := context.WithCancel(ctx)
		return func() error {
				logger.Infof("cron runner started")
				return s.Cron.Run(ctx)
			}, func(err error) {
				cancel()
			}, nil
	}

	return nil, nil, nil
}

func (s serveIn) signalWatch(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(ctx)
	return func() error {
			select {
			case n := <-sig:
				logger.Errf("signal received: %s", n)
			case <-ctx.Done():
				logger.Errf(ctx.Err().Error())
			}
			return nil
		}, func(err error) {
			signal.Stop(sig)
			cancel()
		}, nil
}

func newServeCmd(s serveIn) *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long:  `Start the gRPC server, HTTP server, and cron job runner.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				g run.Group
				l = logging.WithLevel(s.Logger)
			)

			for _, m := range s.Container.Modules() {
				l.Debugf("load module: %T", m)
			}

			// Polyfill missing dependencies
			setDefaultLifecycles(&s)

			// Add serve and signalWatch
			serves := []runGroupFunc{
				s.httpServe,
				s.grpcServe,
				s.cronServe,
				s.signalWatch,
			}

			for _, serve := range serves {
				execute, interrupt, err := serve(cmd.Context(), l)
				if err != nil {
					return err
				}
				if execute == nil {
					continue
				}
				g.Add(execute, interrupt)
			}

			// Additional run groups
			applyRunGroup(s.Container, &g)

			if err := g.Run(); err != nil {
				return err
			}

			l.Info("graceful shutdown complete; see you next time :)")
			return nil
		},
	}
	return serveCmd
}

func applyRouter(ctn contract.Container, router *mux.Router) {
	modules := ctn.Modules()
	for i := range modules {
		if p, ok := modules[i].(HTTPProvider); ok {
			p.ProvideHTTP(router)
		}
	}
}

func applyGRPCServer(ctn contract.Container, server *grpc.Server) {
	modules := ctn.Modules()
	for i := range modules {
		if p, ok := modules[i].(GRPCProvider); ok {
			p.ProvideGRPC(server)
		}
	}
}

func applyRunGroup(ctn contract.Container, group *run.Group) {
	modules := ctn.Modules()
	for i := range modules {
		if p, ok := modules[i].(RunProvider); ok {
			p.ProvideRunGroup(group)
		}
		if p, ok := modules[i].(Runnable); ok {
			ctx, cancel := context.WithCancel(context.Background())
			group.Add(func() error {
				return p.Run(ctx)
			}, func(err error) {
				cancel()
			})
		}
	}
}

func applyCron(ctn contract.Container, cron *cron.Cron) {
	modules := ctn.Modules()
	for i := range modules {
		if p, ok := modules[i].(CronProvider); ok {
			p.ProvideCron(cron)
		}
	}
}

func setDefaultLifecycles(s *serveIn) {
	defaultLifecycles := provideLifecycle()
	if s.HTTPServerStart == nil {
		s.HTTPServerStart = defaultLifecycles.HTTPServerStart
	}
	if s.HTTPServerShutdown == nil {
		s.HTTPServerShutdown = defaultLifecycles.HTTPServerShutdown
	}
	if s.GRPCServerStart == nil {
		s.GRPCServerStart = defaultLifecycles.GRPCServerStart
	}
	if s.GRPCServerShutdown == nil {
		s.GRPCServerShutdown = defaultLifecycles.GRPCServerShutdown
	}
}

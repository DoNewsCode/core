package core

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log/level"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DoNewsCode/core/container"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type serveIn struct {
	di.In

	Dispatcher contract.Dispatcher
	Config     contract.ConfigAccessor
	Logger     log.Logger
	Container  contract.Container
	HTTPServer *http.Server `optional:"true"`
	GRPCServer *grpc.Server `optional:"true"`
	Cron       *cron.Cron   `optional:"true"`
}

func NewServeModule(in serveIn) serveModule {
	return serveModule{
		in,
	}
}

var _ container.CommandProvider = (*serveModule)(nil)

type serveModule struct {
	in serveIn
}

func (s serveModule) ProvideCommand(command *cobra.Command) {
	command.AddCommand(newServeCmd(s.in))
}

type runGroupFunc func(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error)

func (s serveIn) httpServe(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error) {
	if s.Config.Bool("http.disable") {
		return nil, nil, nil
	}

	if s.HTTPServer == nil {
		s.HTTPServer = &http.Server{}
	}
	router := mux.NewRouter()
	if s.Container.ApplyRouter(router) == 0 {
		logger.Info("no http service to apply")
		return nil, nil, nil
	}
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, _ := route.GetPathTemplate()
		level.Debug(logger).Log("service", "http", "path", tpl)
		return nil
	})

	s.HTTPServer.Handler = router

	httpAddr := s.Config.String("http.addr")
	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed start http server")
	}
	return func() error {
			logger.Infof("http service is listening at %s", ln.Addr())
			s.Dispatcher.Dispatch(
				ctx,
				events.Of(OnHTTPServerStart{s.HTTPServer, ln}),
			)
			defer s.Dispatcher.Dispatch(
				ctx,
				events.Of(OnHTTPServerShutdown{s.HTTPServer, ln}),
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
		opts := []grpc.ServerOption{
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		}
		s.GRPCServer = grpc.NewServer(opts...)
	}
	if s.Container.ApplyGRPCServer(s.GRPCServer) == 0 {
		logger.Info("no grpc service to apply")
		return nil, nil, nil
	}

	for module, info := range s.GRPCServer.GetServiceInfo() {
		for _, method := range info.Methods {
			//logger.Debugf("grpc path: %s/%s", module, method.Name)
			level.Debug(logger).Log("service", "grpc", "path", fmt.Sprintf("%s/%s", module, method.Name))
		}
	}

	grpcAddr := s.Config.String("grpc.addr")
	ln, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed start grpc server")
	}
	return func() error {
			logger.Infof("gRPC service is listening at %s", ln.Addr())
			s.Dispatcher.Dispatch(
				ctx,
				events.Of(OnGRPCServerStart{s.GRPCServer, ln}),
			)
			defer s.Dispatcher.Dispatch(
				ctx,
				events.Of(OnGRPCServerShutdown{s.GRPCServer, ln}),
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
		s.Cron = cron.New(cron.WithLogger(cronopts.CronLogAdapter{Logging: s.Logger}))
	}

	if s.Container.ApplyCron(s.Cron) == 0 {
		logger.Info("no cron service to apply")
		return nil, nil, nil
	}
	return func() error {
			logger.Infof("cron runner started")
			s.Cron.Run()
			return nil
		}, func(err error) {
			<-s.Cron.Stop().Done()
		}, nil
}

func (s serveIn) signalWatch(ctx context.Context, logger logging.LevelLogger) (func() error, func(err error), error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	return func() error {
			select {
			case n := <-sig:
				logger.Errf("signal received: %s", n)
			case <-ctx.Done():
				logger.Errf(ctx.Err().Error())
			}
			return nil
		}, func(err error) {
			close(sig)
		}, nil
}

func newServeCmd(s serveIn) *cobra.Command {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long:  `Start the gRPC server, HTTP server, and cron job runner.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var (
				g run.Group
				l = logging.WithLevel(s.Logger)
			)

			// Add serve and signalWatch
			serves := []runGroupFunc{
				s.httpServe,
				s.grpcServe,
				s.cronServe,
				s.signalWatch,
			}

			serveCount := len(serves)

			for _, serve := range serves {
				execute, interrupt, err := serve(cmd.Context(), l)
				if err != nil {
					return err
				}
				if execute == nil {
					serveCount--
					continue
				}
				g.Add(execute, interrupt)
			}

			// Additional run groups
			serveCount += s.Container.ApplyRunGroup(&g)

			// At least one serve: signalWatch. It doesn't need to run alone
			if serveCount <= 1 {
				l.Warn("there are no services to run, please check module")
				return nil
			}

			if err := g.Run(); err != nil {
				return err
			}

			l.Info("graceful shutdown complete; see you next time :)")
			return nil
		},
	}
	return serveCmd
}

package core

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/DoNewsCode/std/pkg/logging"
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

	Config     contract.ConfigAccessor
	Logger     log.Logger
	Container  contract.Container
	HttpServer *http.Server `optional:"true"`
	GrpcServer *grpc.Server `optional:"true"`
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
	newServeCmd(s.in)
}

func newServeCmd(p serveIn) *cobra.Command {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long:  `Start the gRPC server, HTTP server, and cron job runner.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var (
				g run.Group
				l = logging.WithLevel(p.Logger)
			)

			// Start HTTP server
			{
				httpAddr := p.Config.String("http.addr")
				ln, err := net.Listen("tcp", httpAddr)
				if err != nil {
					return errors.Wrap(err, "failed start http server")
				}
				if p.HttpServer == nil {
					p.HttpServer = &http.Server{}
				}
				router := mux.NewRouter()
				p.Container.ApplyRouter(router)
				p.HttpServer.Handler = router
				g.Add(func() error {
					l.Infof("http service is listening at %s", ln.Addr())
					return p.HttpServer.Serve(ln)
				}, func(err error) {
					_ = p.HttpServer.Shutdown(context.Background())
					_ = ln.Close()
				})
			}

			// Start gRPC server
			{
				grpcAddr := p.Config.String("grpc.addr")
				ln, err := net.Listen("tcp", grpcAddr)
				if err != nil {
					return errors.Wrap(err, "failed start grpc server")
				}
				if p.GrpcServer == nil {
					p.GrpcServer = grpc.NewServer()
				}
				p.Container.ApplyGRPCServer(p.GrpcServer)
				g.Add(func() error {
					l.Infof("gRPC service is listening at %s", ln.Addr())
					return p.GrpcServer.Serve(ln)
				}, func(err error) {
					p.GrpcServer.GracefulStop()
					_ = ln.Close()
				})
			}

			// Start cron runner
			{
				if p.Cron == nil {
					p.Cron = cron.New()
				}
				p.Container.ApplyCron(p.Cron)
				g.Add(func() error {
					l.Info("cron runner started")
					p.Cron.Run()
					return nil
				}, func(err error) {
					<-p.Cron.Stop().Done()
				})
			}

			// Graceful shutdown
			{
				sig := make(chan os.Signal, 1)
				signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
				g.Add(func() error {
					select {
					case s := <-sig:
						l.Errf("signal received: %s", s)
					case <-cmd.Context().Done():
						l.Errf(cmd.Context().Err().Error())
					}
					return nil
				}, func(err error) {
					close(sig)
				})
			}

			// Additional run groups
			p.Container.ApplyRunGroup(&g)

			if err := g.Run(); err != nil {
				return err
			}

			l.Infof("graceful shutdown complete; see you next time :)")
			return nil
		},
	}
	return serveCmd
}

package contract

import (
	"github.com/Reasno/ifilter"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type Container interface {
	GetHttpProviders() []func(router *mux.Router)
	GetGrpcProviders() []func(server *grpc.Server)
	GetCloserProviders() []func()
	GetRunProviders() []func(g *run.Group)
	GetModules() ifilter.Collection
	GetCronProviders() []func(crontab *cron.Cron)
	GetCommandProviders() []func(command *cobra.Command)
	AddModule(module interface{})
}

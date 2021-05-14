package contract

import (
	"github.com/Reasno/ifilter"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// Container holds modules.
type Container interface {
	ApplyRouter(router *mux.Router) int
	ApplyGRPCServer(server *grpc.Server) int
	ApplyCron(crontab *cron.Cron) int
	ApplyRunGroup(g *run.Group) int
	ApplyRootCommand(command *cobra.Command)
	Shutdown()
	Modules() ifilter.Collection
	AddModule(module interface{})
}

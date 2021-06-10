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
	ApplyRouter(router *mux.Router)
	ApplyGRPCServer(server *grpc.Server)
	ApplyCron(crontab *cron.Cron)
	ApplyRunGroup(g *run.Group)
	ApplyRootCommand(command *cobra.Command)
	Shutdown()
	Modules() ifilter.Collection
	AddModule(module interface{})
}

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
	Shutdown()
	ApplyRunGroup(g *run.Group)
	Modules() ifilter.Collection
	ApplyCron(crontab *cron.Cron)
	ApplyRootCommand(command *cobra.Command)
	AddModule(module interface{})
}

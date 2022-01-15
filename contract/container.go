package contract

import (
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// Container holds modules.
type Container interface {
	ApplyRouter(router *mux.Router)
	ApplyGRPCServer(server *grpc.Server)
	ApplyRunGroup(g *run.Group)
	ApplyRootCommand(command *cobra.Command)
	Shutdown()
	Modules() []interface{}
	AddModule(module interface{})
}

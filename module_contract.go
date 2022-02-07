package core

import (
	"github.com/DoNewsCode/core/cron"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// CronProvider provides cron jobs.
type CronProvider interface {
	ProvideCron(cron *cron.Cron)
}

// CommandProvider provides cobra.Command.
type CommandProvider interface {
	ProvideCommand(command *cobra.Command)
}

// HTTPProvider provides http services.
type HTTPProvider interface {
	ProvideHTTP(router *mux.Router)
}

// GRPCProvider provides gRPC services.
type GRPCProvider interface {
	ProvideGRPC(server *grpc.Server)
}

// CloserProvider provides a shutdown function that will be called when service exits.
type CloserProvider interface {
	ProvideCloser()
}

// RunProvider provides a runnable actor. Use it to register any server-like
// actions. For example, kafka consumer can be started here.
type RunProvider interface {
	ProvideRunGroup(group *run.Group)
}

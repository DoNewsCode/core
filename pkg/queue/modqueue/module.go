package modqueue

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Module struct {
	Factory *DispatcherFactory
}

func New(factory *DispatcherFactory) Module {
	return Module{Factory: factory}
}

func (m Module) ProvideCommand(command *cobra.Command) {
	var queueName string
	var channels []string
	queueCmd := &cobra.Command{
		Use:   "queue reload|flush [-q queue] [-c channels]...",
		Short: "reload or flush the queue",
		Long:  "reload or flush the channels provided by the queue driver",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "reload" {
				queueDispatcher, _ := m.Factory.Make(queueName)
				driver := queueDispatcher.Driver()
				for _, ch := range channels {
					if _, err := driver.Reload(command.Context(), ch); err != nil {
						return errors.Wrap(err, "queue reload command")
					}
				}
				return nil
			}
			if args[0] == "flush" {
				queueDispatcher, _ := m.Factory.Make(queueName)
				driver := queueDispatcher.Driver()
				for _, ch := range channels {
					if err := driver.Flush(command.Context(), ch); err != nil {
						return errors.Wrap(err, "queue flush command")
					}
				}
				return nil
			}
			return fmt.Errorf("invalid argument %s, want flush or reload", args[0])
		},
	}
	queueCmd.Flags().StringVarP(&queueName, "queue", "q", "default", "the queue name")
	queueCmd.Flags().StringSliceVarP(&channels, "channels", "c", []string{"timeout", "failed"}, "the queue name")
	command.AddCommand(queueCmd)
}

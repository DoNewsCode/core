package queue

import (
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
	flushCmd := &cobra.Command{
		Use:   "flush [-q queue] [-c channels]...",
		Short: "flush the timeout or failed events",
		Long:  "flush the timeout or failed events stored in the queue.",
		RunE: func(cmd *cobra.Command, args []string) error {
			queueDispatcher, _ := m.Factory.Make(queueName)
			driver := queueDispatcher.Driver()
			for _, ch := range channels {
				if err := driver.Flush(command.Context(), ch); err != nil {
					return errors.Wrap(err, "queue flush command")
				}
			}
			return nil
		},
	}
	reloadCmd := &cobra.Command{
		Use:   "reload [-q queue] [-c channels]...",
		Short: "reload the timeout or failed events",
		Long:  "move the timeout or failed events to the waiting channel, giving them one more chance to be processed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			queueDispatcher, _ := m.Factory.Make(queueName)
			driver := queueDispatcher.Driver()
			for _, ch := range channels {
				if _, err := driver.Reload(command.Context(), ch); err != nil {
					return errors.Wrap(err, "queue reload command")
				}
			}
			return nil
		},
	}
	queueCmd := &cobra.Command{
		Use:   "queue",
		Short: "manage queues",
		Long:  "manage queues, such as reloading failed command.",
	}
	queueCmd.PersistentFlags().StringVarP(&queueName, "queue", "q", "default", "the queue name")
	queueCmd.PersistentFlags().StringSliceVarP(&channels, "channels", "c", []string{"timeout", "failed"}, "the queue name")
	queueCmd.AddCommand(reloadCmd, flushCmd)
	command.AddCommand(queueCmd)
}

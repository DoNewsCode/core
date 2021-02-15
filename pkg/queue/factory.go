package queue

import (
	"context"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
)

// Dispatcher is the key of *QueueableDispatcher in the dependencies graph. Used as a type hint for injection.
type Dispatcher interface {
	contract.Dispatcher
	Consume(ctx context.Context) error
}

// DispatcherMaker is the key of *DispatcherFactory in the dependencies graph. Used as a type hint for injection.
type DispatcherMaker interface {
	Make(string) (*QueueableDispatcher, error)
}

var _ Dispatcher = (*QueueableDispatcher)(nil)
var _ DispatcherMaker = (*DispatcherFactory)(nil)

// DispatcherFactory is a factory for *QueueableDispatcher
type DispatcherFactory struct {
	*async.Factory
}

// Make returns a QueueableDispatcher by the given name. If it has already been created under the same name,
// the that one will be returned.
func (s *DispatcherFactory) Make(name string) (*QueueableDispatcher, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*QueueableDispatcher), nil
}

type queueConf struct {
	Parallelism                    int `yaml:"parallelism" json:"parallelism"`
	CheckQueueLengthIntervalSecond int `yaml:"checkQueueLengthIntervalSecond" json:"checkQueueLengthIntervalSecond"`
}

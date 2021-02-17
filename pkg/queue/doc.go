// Package queue provides a persistent queue implementation that interplays with the Dispatcher in the contract
// package.
//
// It is recommended to read documentation on the events package before getting started on the queue package.
//
// Introduction
//
// Queues in go is not as prominent as in some other languages, since go excels at concurrency. However,
// the persistent
// queue can still offer some benefit missing from the native mechanism, say go channels.
// The queued job won't be lost
// even if the system shutdown. In other word, it means jobs can be retried until success. Plus, it is also
// possible to queue the execution of a particular job until a lengthy period of time. Useful when you need to
// implement "send email after 30 days" type of event handler.
//
// Simple Usage
//
// To convert any valid event to a persisted event, use:
//
//  pevent := queue.Persist(event)
//
// Like the event package, you don't have to use this helper. Manually create a queueable event by implementing this
// interface on top of the normal event interface:
//
//  type persistent interface {
//    Defer() time.Duration
//    Decorate(s *PersistedEvent)
//  }
//
// The PersistentEvent passed to the Decorate method contains the tunable configuration such as maximum retries.
//
// No matter how you create a persisted event, to fire it, send it though a dispatcher. The normal dispatcher in the
// events package won't work, as a queue implementation is required. Luckily, it is deadly simple to convert a standard
// dispatcher to a queue.Dispatcher.
//
//  queueableDispatcher := queue.WithQueue(&events.SyncDispatcher, &queue.RedisDriver{})
//  queueableDispatcher.dispatch(pevent)
//
// As you see, how the queue persist the events is subject to the underlying driver. The default driver bundled in this
// package is the redis driver.
//
// Once the persisted event are stored in the external storage, a goroutine should consume them and pipe the
// reconstructed event to the listeners. This is done by calling the Consume method of queue.Dispatcher
//
//  go dispatcher.Consume(context.Background())
//
// There is no difference between listeners for normal event and listeners for persisted event. They can be
// used interchangeably. But note if a event is retryable, it is your responsibility to ensure the idempotency.
// Also, be aware if a persisted event have many listeners, the event is up to retry when any of the listeners fail.
//
// Integrate
//
// The queue package exports configuration in this format:
//
//  queue:
//    default:
//      parallelism: 3
//      checkQueueLengthIntervalSecond: 15
//
// While manually constructing the queue.Dispatcher is absolutely feasible, users can use the bundled dependency provider
// without breaking a sweat. Using this approach, the life cycle of consumer goroutine will be managed
// automatically by the core.
//
//  var c *core.C
//  c.AddDependencyFunc(queue.ProvideDispatcher)
//
// A module is also bundled, providing the queue command.
//
//  c.AddModuleFunc(queue.New)
//
// Sometimes there are valid reasons to use more than one queue. Each dispatcher however is bounded to only one queue.
// To use multiple queues, multiple dispatchers are required. Inject
// queue.DispatcherMaker to factory a dispatcher with a specific name.
//
//  c.Invoke(function(maker queue.DispatcherMaker) {
//    dispatcher, err := maker.Make("default")
//    // see examples for details
//  })
//
// Events
//
// When an attempt to execute the event handler failed, two kinds of event will be fired. If the failed event can be
// retried, "queue.RetryingEvent" will be fired. If not, "queue.AbortedEvent" will be fired.
//
// Metrics
//
// To gain visibility on how the length of the queue, inject a gauge into the core and alias it to queue.Gauge. The
// queue length of the all internal queues will be periodically reported to metrics collector (Presumably Prometheus).
//
//  c.AddDependencyFunc(func(appName contract.AppName, env contract.Env) queue.Gauge {
//    return prometheus.NewGaugeFrom(
//      stdprometheus.GaugeOpts{
//        Namespace: appName.String(),
//        Subsystem: env.String(),
//        Name:      "queue_length",
//        Help:      "The gauge of queue length",
//      }, []string{"name", "channel"},
//    )
//  })
package queue

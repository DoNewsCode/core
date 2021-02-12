// Package queue provides a persistent queue implementation that interplays with the event QueueableDispatcher in the contract
// package.
//
// Queues in go is not as prominent as in some other languages, since go excels at concurrency. However,
// the persistent
// queue can still offer some benefit missing from the native mechanism, say go channels.
// The queued job won't be lost
// even if the system shutdown. In other word, it means jobs can be retried until success. Plus, it is also
// possible to queue the execution of a particular job until a lengthy period of time. Useful when you needs to
// implement "cancel vip after 30 days" type of event handler.
package queue

/*
Package events provides a simple and effective implementation of event system.

Event is a battle proven way to decoupling services. Package event calls event
listeners in a synchronous, sequential execution. The synchronous listener is
only a "go" away from an asynchronous handler, but asynchronous listener can not
be easily made synchronous.

The event listeners can also be used as hooks. If the event data is a pointer type,
listeners may alter the data. This enables plugin/addon style decoupling.

Note: Package event focus on events within the system, not events outsource to
eternal system. For that, use a message queue like kafka.
*/
package events

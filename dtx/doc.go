/*
Package dtx contains common utilities in the context of distributed transaction.

Context Passing

It is curial for all parties in the distributed transaction to share an
transaction id. This package provides utility to pass this id across services.

	HTTPToContext() http.RequestFunc
	ContextToHTTP() http.RequestFunc
	GRPCToContext() grpc.ServerRequestFunc
	ContextToGRPC() grpc.ClientRequestFunc

Idempotency

Certain operations will be retried by the client more than once. A middleware is
provided for the server to shield against repeated request in the same
transaction.

	func MakeIdempotence(s Oncer) endpoint.Middleware

Lock

Certain resource in transaction cannot be concurrently accessed. A middleware is
provided to lock such resources.

	func MakeLock(l Locker) endpoint.Middleware

Allow Null Compensation and Prevent Resource Suspension

Transaction participants may receive the compensation
order before performing normal operations due to network exceptions. In this
case, null compensation is required.

If the forward operation arrives later than the compensating operation due to
network exceptions, the forward operation must be discarded. Otherwise, resource
suspension occurs.

	func MakeAttempt(s Sequencer) endpoint.Middleware
	func MakeCancel(s Sequencer) endpoint.Middleware

*/
package dtx

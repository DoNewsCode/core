/*
Package kitkafka provides a kafka transport for go kit.

Introduction

Go kit has some great properties, such as allowing multiple transport to be used
simultaneously. Sadly it limits itself to only support RPCs. In real projects
with many decoupled component, messaging is an inevitable path we must go down.

Go kit models the RPCs as:

	func(context.Context, request interface{}) (response interface{}, err error)

Package kitkafka treat messaging as a special case RPC, where the response is always ignored.
By using the same model, package kafka brings all
*/
package kitkafka
/*
Package observability provides a tracer and a histogram to measure all incoming
requests to the system.

Introduction

Observability is crucial to the stability of the system. The three pillars of the observabilities
consist of logging, tracing and metrics. Since logging is provided in Package logging, this package
mainly focus on tracing and metrics.

Integration

Add the observabilities to core:

	var c *core.C = core.New()
	c.provide(observability.Providers())

See example for usage.
*/
package observability

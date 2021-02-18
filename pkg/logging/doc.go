/*
Package logging provides a kitlog compatible logger.

This package is mostly a thin wrapper around kitlog
(http://github.com/go-kit/kit/log). kitlog provides a minimalist, contextual,
fully composable logger. However it is too unopinionated, hence requiring some
efforts and coordination to set up a good practise.

Integration

Package logging is bundled in core. Enable logging as dependency by calling:

	var c *core.C = core.New()
	c.AddCoreDependencies()

See example for usage.
*/
package logging

/*
Package contract defines a set of common interfaces for all packages in this repository.

Note the purpose of this packages is to document what contract the libraries for
package core should agree upon.

In general, use a centralized package for contracts is an anti-pattern in go as
it prevents progressive upgrade. It is recommended to redeclare them in each
individual package, unless there is a reason not to, eg. import cycle.

Package contract should be considered a closed set. Adding new interfaces to this
package is strongly discouraged.
*/
package contract

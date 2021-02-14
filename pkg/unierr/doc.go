// Package unierr presents an unification error model between gRPC transport and HTTP transport,
// and between server and client.
//
// It is modeled after the gRPC status.
//
// To create an not found error with a custom message:
//
//  unierr.New(codes.NotFound, "some stuff is missing")
//
// To wrap an existing error:
//
//  unierr.Wrap(err, codes.NotFound)
//
// See example for detailed usage.
package unierr

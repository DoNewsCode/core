package logging_test

import (
	"context"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log/level"
)

func Example_minimal() {
	logger := logging.NewLogger("json")
	logger.Log("foo", "bar")
	// Output:
	// {"foo":"bar"}
}

func Example_level() {
	logger := logging.NewLogger("json")
	level.Info(logger).Log("foo", "bar")
	// Output:
	// {"foo":"bar","level":"info"}
}

func ExampleWithLevel() {
	logger := logging.NewLogger("json")
	levelLogger := logging.WithLevel(logger)
	levelLogger.Info("hello")
	// Output:
	// {"level":"info","msg":"hello"}
}

func ExampleWithContext() {
	ctx := context.WithValue(context.Background(), contract.IpKey, "127.0.0.1")
	ctx = context.WithValue(ctx, contract.TransportKey, "http")
	ctx = context.WithValue(ctx, contract.RequestUrlKey, "/example")
	logger := logging.NewLogger("json")
	ctxLogger := logging.WithContext(logger, ctx)
	ctxLogger.Log("foo", "bar")
	// Output:
	// {"clientIp":"127.0.0.1","foo":"bar","requestUrl":"/example","transport":"http"}
}

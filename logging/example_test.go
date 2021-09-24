package logging_test

import (
	"context"

	"github.com/DoNewsCode/core/ctxmeta"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log/level"
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
	// {"caller":"example_test.go:28","level":"info","msg":"hello"}
}

func ExampleWithContext() {
	bag, ctx := ctxmeta.Inject(context.Background())
	bag.Set("clientIp", "127.0.0.1")
	bag.Set("requestUrl", "/example")
	bag.Set("transport", "http")
	logger := logging.NewLogger("json")
	ctxLogger := logging.WithContext(logger, ctx)
	ctxLogger.Log("foo", "bar")
	// Output:
	// {"clientIp":"127.0.0.1","foo":"bar","requestUrl":"/example","transport":"http"}
}

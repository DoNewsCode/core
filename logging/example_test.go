package logging_test

import (
	"context"
	"fmt"

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
	// {"caller":"example_test.go:29","level":"info","msg":"hello"}
}

func Example_sprintf() {
	logger := logging.NewLogger("json")
	// Set log level to info
	logger = level.NewFilter(logger, level.AllowInfo())
	levelLogger := logging.WithLevel(logger)

	// Let's try to log some debug messages. They are filtered by log level, so you should see no output.
	// The cost of fmt.Sprintf is paid event if the log is filtered out. This sometimes can be a huge performance downside.
	levelLogger.Debugw("record some data", "data", fmt.Sprintf("%+v", []int{1, 2, 3}))
	// Or better, we can use logging.Sprintf to avoid the cost if the log is not actually written to the output.
	levelLogger.Debugw("record some data", "data", logging.Sprintf("%+v", []int{1, 2, 3}))

	// Output:
	//
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

package logging

import "fmt"

type sprintf struct {
	format string
	args   []interface{}
}

// String returns the formatted string value using fmt.Sprintf.
func (s sprintf) String() string {
	return fmt.Sprintf(s.format, s.args...)
}

// Sprintf returns a log entry that is formatted using fmt.Sprintf just before
// writing to the output. This is more desirable than using fmt.Sprintf from the
// caller's end because the cost of formatting can be avoided if the log is
// filtered, for example, by log level.
func Sprintf(format string, args ...interface{}) fmt.Stringer {
	return sprintf{format: format, args: args}
}

type sprint struct {
	args []interface{}
}

// String returns the formatted string value using fmt.Sprint.
func (s sprint) String() string {
	return fmt.Sprint(s.args...)
}

// Sprint returns a log entry that is formatted using fmt.Sprint just before
// writing to the output. This is more desirable than using fmt.Sprint from the
// caller's end because the cost of formatting can be avoided if the log is
// filtered, for example, by log level.
func Sprint(args ...interface{}) fmt.Stringer {
	return sprint{args: args}
}

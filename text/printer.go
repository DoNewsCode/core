// Package text provides utilities to generate textual output.
package text

import "fmt"

// Printer is an interface for i18n and l10n manipulations. For example, an Chinese printer can
// translate the message and arguments to its Chinese counterpart.
type Printer interface {
	Sprintf(msg string, val ...any) string
}

// BasePrinter is the default printer for common use. It uses fmt.Sprintf underneath.
type BasePrinter struct{}

// Sprintf formats according to a format specifier and returns the resulting string.
func (BasePrinter) Sprintf(msg string, val ...any) string {
	return fmt.Sprintf(msg, val...)
}

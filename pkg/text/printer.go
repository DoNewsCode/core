package text

import "fmt"

type Printer interface {
	Sprintf(msg string, val ...interface{}) string
}

type BasePrinter struct {}

func (BasePrinter) Sprintf(msg string, val ...interface{}) string {
	return fmt.Sprintf(msg, val...)
}

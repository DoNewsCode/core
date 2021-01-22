package contract

type Printer interface {
	Sprintf(msg string, val ...interface{}) string
}

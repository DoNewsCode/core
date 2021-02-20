package contract

// A Printer models a i18n translator.
type Printer interface {
	Sprintf(msg string, val ...interface{}) string
}

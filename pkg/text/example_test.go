package text_test

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/text"
)

func ExampleBasePrinter_Sprintf() {
	var printer text.BasePrinter
	fmt.Println(printer.Sprintf("hello %s", "go"))
	// Output:
	// hello go
}

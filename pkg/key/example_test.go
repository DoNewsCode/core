package key_test

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/key"
)

func Example() {
	keyer := key.NewManager("module", "foo")
	fmt.Println(keyer.Spread())
	// Output:
	// [module foo]
}

func ExampleWith() {
	keyer := key.NewManager("module", "foo")
	fmt.Println(key.With(keyer, "service", "bar").Spread())
	// Output:
	// [module foo service bar]
}

func ExampleKeepOdd() {
	keyer := key.NewManager("module", "foo", "service", "bar")
	fmt.Println(key.KeepOdd(keyer).Spread())
	// Output:
	// [foo bar]
}

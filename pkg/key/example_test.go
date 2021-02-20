package key_test

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/key"
)

func Example() {
	keyer := key.New("module", "foo")
	fmt.Println(keyer.Spread())
	// Output:
	// [module foo]
}

func ExampleWith() {
	keyer := key.New("module", "foo")
	fmt.Println(key.With(keyer, "service", "bar").Spread())
	// Output:
	// [module foo service bar]
}

func ExampleKeepOdd() {
	keyer := key.New("module", "foo", "service", "bar")
	fmt.Println(key.KeepOdd(keyer).Spread())
	// Output:
	// [foo bar]
}

func ExampleKeyManager_Key() {
	keyer := key.New("module", "foo", "service", "bar")
	fmt.Println(keyer.Key("."))
	// Output:
	// module.foo.service.bar
}

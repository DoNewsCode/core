package events

import (
	"fmt"
	"testing"
)

type TestE struct {
}

func TestEvent(t *testing.T) {
	testE := Of(TestE{})
	fmt.Println(testE.Type())
}

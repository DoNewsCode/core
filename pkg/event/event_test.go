package event

import (
	"fmt"
	"testing"
)

type TestE struct {
}

func TestEvent(t *testing.T) {
	testE := NewEvent(TestE{})
	fmt.Println(testE.Type())
}

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_AddModule(t *testing.T) {
	cases := []struct {
		name    string
		module  interface{}
		asserts func(t *testing.T, container Container)
	}{
		{
			"any",
			"foo",
			func(t *testing.T, container Container) {
				assert.Contains(t, container.Modules(), "foo")
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var container Container
			container.AddModule(c.module)
			c.asserts(t, container)
		})
	}
}

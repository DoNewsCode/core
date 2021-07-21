package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := provideDefaultConfig()
	for _, c := range conf {
		if c.Validate != nil {
			err := c.Validate(c.Data)
			assert.NoError(t, err)
		}
	}
}

func TestDefaultConfig_invalid(t *testing.T) {
	conf := provideDefaultConfig()

	t.Run("empty", func(t *testing.T) {
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{})
				assert.Error(t, err)
			}
		}
	})

	t.Run("invalid http addr", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"http": map[string]interface{}{
						"addr":    "aaa",
						"disable": false,
					},
				})
				assert.Error(t, err)
			}
		}
	})

	t.Run("invalid grpc addr", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"grpc": map[string]interface{}{
						"addr":    "aaa",
						"disable": false,
					},
				})
				assert.Error(t, err)
			}
		}
	})

	t.Run("disabled transport http", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"http": map[string]interface{}{
						"addr":    "aaa",
						"disable": true,
					},
				})
				if err == nil {
					return
				}
			}
		}
		t.Error("disabled transport should not have addr requirement")
	})

	t.Run("disabled transport grpc", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"grpc": map[string]interface{}{
						"addr":    "aaa",
						"disable": true,
					},
				})
				if err == nil {
					return
				}
			}
		}
		t.Error("disabled transport should not have addr requirement")
	})

	t.Run("wrong type", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"grpc": map[string]interface{}{
						"disable": "",
					},
				})
				assert.Error(t, err)
			}
		}
	})

	t.Run("wrong env", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"env": "bar",
				})
				assert.Error(t, err)
			}
		}
	})

	t.Run("wrong app", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"app": 1,
				})
				assert.Error(t, err)
			}
		}
	})

	t.Run("wrong log level", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"level": map[string]interface{}{
						"format": "json",
						"level":  "all",
					},
				})
				assert.Error(t, err)
			}
		}
	})

	t.Run("wrong log format", func(t *testing.T) {
		conf := provideDefaultConfig()
		for _, c := range conf {
			if c.Validate != nil {
				err := c.Validate(map[string]interface{}{
					"level": map[string]interface{}{
						"format": "foo",
						"level":  "debug",
					},
				})
				assert.Error(t, err)
			}
		}
	})
}

func TestDefaultConfig_network(t *testing.T) {
	conf := provideDefaultConfig()
	for _, c := range conf {
		if c.Validate != nil {
			err := c.Validate(map[string]interface{}{
				"http": map[string]interface{}{
					"addr":    "aaa",
					"disable": false,
				},
			})
			assert.Error(t, err)
		}
	}
}

package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_UnmarshalJSON(t *testing.T) {
	var cases = []struct {
		name     string
		value    string
		expected Duration
	}{
		{
			"simple",
			`"5s"`,
			Duration{5 * time.Second},
		},
		{
			"float",
			`65000000000.0`,
			Duration{5*time.Second + time.Minute},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v := Duration{}
			json.Unmarshal([]byte(c.value), &v)
			assert.Equal(t, c.expected, v)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	var cases = []struct {
		name     string
		value    Duration
		expected string
	}{
		{
			"simple",
			Duration{5 * time.Second},
			`"5s"`,
		},
		{
			"complex",
			Duration{5*time.Second + time.Minute},
			`"1m5s"`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(c.value)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, string(data))
		})
	}
}

package mysqlstore

import (
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
)

func Test_getRetention(t *testing.T) {
	cases := []struct {
		name     string
		conf     configuration
		expected config.Duration
	}{
		{
			"zero value",
			configuration{
				Connection:      "",
				Retention:       config.Duration{},
				CleanupInterval: config.Duration{},
			},
			defaultConfig.Retention,
		},
		{
			name: "non zero value",
			conf: configuration{
				Connection:      "",
				Retention:       config.Duration{Duration: time.Second},
				CleanupInterval: config.Duration{},
			},
			expected: config.Duration{Duration: time.Second},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.expected, c.conf.getRetention())
		})
	}
}

func Test_getCleanupInterval(t *testing.T) {
	cases := []struct {
		name     string
		conf     configuration
		expected config.Duration
	}{
		{
			"zero value",
			configuration{
				Connection:      "",
				Retention:       config.Duration{},
				CleanupInterval: config.Duration{},
			},
			defaultConfig.CleanupInterval,
		},
		{
			name: "non zero value",
			conf: configuration{
				Connection:      "",
				Retention:       config.Duration{Duration: time.Second},
				CleanupInterval: config.Duration{Duration: time.Second},
			},
			expected: config.Duration{Duration: time.Second},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.expected, c.conf.getCleanupInterval())
		})
	}
}

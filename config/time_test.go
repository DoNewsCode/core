package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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
			v1 := Duration{}
			yaml.Unmarshal([]byte(c.value), &v1)
			assert.Equal(t, c.expected, v1)
			v2 := Duration{}
			json.Unmarshal([]byte(c.value), &v2)
			assert.Equal(t, c.expected, v2)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	var cases = []struct {
		name         string
		value        interface{}
		expectedJSON string
		expectedYaml string
	}{
		{
			"simple",
			Duration{5 * time.Second},
			`"5s"`,
			"5s\n",
		},
		{
			"complex",
			Duration{5*time.Second + time.Minute},
			`"1m5s"`,
			"1m5s\n",
		},
		{
			"wrapped",
			struct{ D Duration }{Duration{5*time.Second + time.Minute}},
			`{"D":"1m5s"}`,
			"d: 1m5s\n",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(c.value)
			assert.NoError(t, err)
			assert.Equal(t, c.expectedJSON, string(data))
			data, err = yaml.Marshal(c.value)
			assert.Equal(t, c.expectedYaml, string(data))
		})
	}
}

func TestDuration_IsZero(t *testing.T) {
	tests := []struct {
		name string
		val  time.Duration
		want bool
	}{
		{">0", 1, false},
		{"<0", -1, false},
		{"=0", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Duration{tt.val}
			if got := d.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuration_UnmarshalText(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("1s"))
	assert.NoError(t, err)
	assert.Equal(t, Duration{time.Second}, d)
}

func TestDuration_MarshalText(t *testing.T) {
	d := Duration{time.Second}
	b, err := d.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("1s"), b)
}

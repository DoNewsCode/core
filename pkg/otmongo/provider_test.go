package otmongo

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func TestNewMongoFactory(t *testing.T) {
	factory, cleanup, err := NewMongoFactory(MongoParam{
		In: dig.In{},
		Conf: config.MapAdapter{"mongo": map[string]struct{ Uri string }{
			"default": {
				Uri: "mongodb://127.0.0.1:27017",
			},
			"alternative": {
				Uri: "mongodb://127.0.0.1:27017",
			},
		}},
		Tracer: nil,
	})
	assert.NoError(t, err)
	assert.Len(t, factory.db, 2)
	assert.NotNil(t, cleanup)
}

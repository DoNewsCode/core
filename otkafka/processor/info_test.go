package processor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInfo(t *testing.T) {
	i := Info{
		Name:         "",
		ReadWorker:   0,
		BatchWorker:  0,
		BatchSize:    0,
		HandleWorker: 0,
		ChanSize:     0,
	}
	assert.Equal(t, i.name(), "default")
	assert.Equal(t, i.readWorker(), 1)
	assert.Equal(t, i.batchWorker(), 1)
	assert.Equal(t, i.batchSize(), 1)
	assert.Equal(t, i.handleWorker(), 1)
	assert.Equal(t, i.chanSize(), 100)

	j := Info{
		Name:         "test",
		ReadWorker:   2,
		BatchWorker:  2,
		BatchSize:    10,
		HandleWorker: 2,
		ChanSize:     10,
	}
	assert.Equal(t, j.name(), j.Name)
	assert.Equal(t, j.readWorker(), j.ReadWorker)
	assert.Equal(t, j.batchWorker(), j.BatchWorker)
	assert.Equal(t, j.batchSize(), j.BatchSize)
	assert.Equal(t, j.handleWorker(), j.HandleWorker)
	assert.Equal(t, j.chanSize(), j.ChanSize)
}

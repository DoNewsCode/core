package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	i := Info{
		Name:         "",
		ReadWorker:   0,
		BatchWorker:  0,
		BatchSize:    0,
		HandleWorker: 0,
		ChanSize:     0,
		AutoCommit:   false,
	}
	assert.Equal(t, "default", i.name())
	assert.Equal(t, 1, i.readWorker())
	assert.Equal(t, 1, i.batchWorker())
	assert.Equal(t, 1, i.batchSize())
	assert.Equal(t, 1, i.handleWorker())
	assert.Equal(t, 100, i.chanSize())
	assert.Equal(t, false, i.autoCommit())

	j := Info{
		Name:         "test",
		ReadWorker:   2,
		BatchWorker:  2,
		BatchSize:    10,
		HandleWorker: 2,
		ChanSize:     10,
		AutoCommit:   true,
	}
	assert.Equal(t, j.Name, j.name())
	assert.Equal(t, j.ReadWorker, j.readWorker())
	assert.Equal(t, j.BatchWorker, j.batchWorker())
	assert.Equal(t, j.BatchSize, j.batchSize())
	assert.Equal(t, j.HandleWorker, j.handleWorker())
	assert.Equal(t, j.ChanSize, j.chanSize())
	assert.Equal(t, j.AutoCommit, j.autoCommit())
}

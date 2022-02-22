package dag

import (
	"context"
	"strconv"
	"sync"

	"github.com/DoNewsCode/core/logging"

	"github.com/go-kit/log"
	"golang.org/x/sync/errgroup"
)

// VertexID is the identifier of Vertex in a directed acyclic graph.
type VertexID int

// Edges is a group of edges that are connected to a vertex.
type Edges [][]VertexID

type vertex struct {
	id       int
	name     string
	logger   log.Logger
	children []*vertex
	parents  []*vertex
	work     func(ctx context.Context) error
	done     chan struct{}
	once     sync.Once
}

func newVertex(id int, work func(ctx context.Context) error, options ...VertexOption) *vertex {
	vertex := &vertex{
		id:   id,
		work: work,
		done: make(chan struct{}),
	}
	for _, f := range options {
		f(vertex)
	}
	if vertex.name == "" {
		vertex.name = "vertex-" + strconv.Itoa(id)
	}
	return vertex
}

func (v *vertex) execute(ctx context.Context) error {
	if v.logger != nil {
		v.logger.Log("msg", logging.Sprintf("started to execute vertex %s", v.name))
		defer v.logger.Log("msg", logging.Sprintf("finished executing vertex %s", v.name))
	}
	for i := range v.parents {
		select {
		case <-v.parents[i].done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := v.work(ctx); err != nil {
		return err
	}
	close(v.done)

	errGroup, ctx := errgroup.WithContext(ctx)
	for _, c := range v.children {
		c := c
		c.once.Do(func() {
			errGroup.Go(func() error {
				return c.execute(ctx)
			})
		})
	}
	return errGroup.Wait()
}

func (v *vertex) addChild(child *vertex) {
	v.children = append(v.children, child)
	child.parents = append(child.parents, v)
}

func (v *vertex) removeChild(child *vertex) {
	for _, c := range v.children {
		if c.id == child.id {
			v.children = append(v.children[:c.id], v.children[c.id+1:]...)
			break
		}
	}
}

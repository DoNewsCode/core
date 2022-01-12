// Package dag offers a simple, pure in-memory job scheduler based on Directed
// Acyclic graph. Most common use cases are to schedule a bunch of interconnected
// job in a cron or cli command.
//
// Each vertex stands for an arbitrary function to be scheduled
// and the edge between them describes their dependency. The scheduler will run
// each vertex in an independent goroutine as soon as all its dependencies are
// finished. Vertexes with no direct dependency may be scheduled concurrently.
// The scheduler will not run any vertex twice.
//
// If a vertex returns an error or if the dag context is canceled, the scheduler
// will prevent any subsequent vertexes from scheduling, cancel all vertex level
// contexts and return to the caller immediately.
package dag

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// DAG is a directed acyclic graph designed for job scheduling.
type DAG struct {
	vertexes []*vertex
}

// New creates a new DAG instance.
func New() *DAG {
	return &DAG{}
}

// AddVertex adds a vertex to the dag. AddVertex is not concurrent safe. All
// vertexes and edges are expected to be added synchronously before calling Run.
func (d *DAG) AddVertex(work func(ctx context.Context) error, option ...VertexOption) VertexID {
	id := len(d.vertexes)
	d.vertexes = append(d.vertexes, newVertex(id, work, option...))
	return VertexID(id)
}

// AddEdge adds an edge to the dag. AddEdge is not concurrent safe. All vertexes
// and edges are expected to be added synchronously before calling Run.
//
// If the new edge leads to a cycle, AddEdge will return error.
func (d *DAG) AddEdge(from, to VertexID) error {
	if len(d.vertexes) <= int(from) || from < 0 {
		return errors.Errorf("invalid vertex id %d", from)
	}
	if to < 0 || len(d.vertexes) <= int(to) {
		return errors.Errorf("invalid vertex id %d", to)
	}

	d.vertexes[from].addChild(d.vertexes[to])
	if yes, edges := IsAcyclic(d); !yes {
		d.vertexes[from].removeChild(d.vertexes[to])
		return errors.Errorf("dag is not acyclic: %s", d.fmtEdges(edges))
	}
	return nil
}

// Run runs the dag. Vertexes with no dependency will be scheduled concurrently
// while the inked vertexes will be scheduled sequentially. The Scheduler
// optimizes the execution path so that the overall dag execution time is
// minimized.
//
// If a vertex returns an error or if the dag context is canceled, the scheduler
// will prevent any subsequent vertexes from scheduling, cancel all vertex level
// contexts and return to the caller immediately.
//
// One of the ways for parent vertexes to pass results to child vertexes (or the
// dag caller) is to store the results in context with the help of package
// ctxmeta. See example.
func (d *DAG) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	for _, v := range d.vertexes {
		v := v
		if len(v.parents) == 0 {
			errGroup.Go(func() error {
				return v.execute(ctx)
			})
		}
	}
	return errGroup.Wait()
}

func (d *DAG) fmtEdges(edges []int) interface{} {
	var s []string
	for _, edge := range edges {
		if d.vertexes[edge].name != "" {
			s = append(s, fmt.Sprintf("%d: %s", edge, d.vertexes[edge].name))
		} else {
			s = append(s, fmt.Sprintf("%d", edge))
		}
	}
	return strings.Join(s, " -> ")
}

// order returns the total number of nodes in the graph
func (d *DAG) order() int {
	return len(d.vertexes)
}

// edgesFrom returns a list of integers that each
// represents a node that has an edge from node u.
func (d *DAG) edgesFrom(u int) []int {
	var edges []int
	for _, c := range d.vertexes[u].children {
		edges = append(edges, c.id)
	}
	return edges
}

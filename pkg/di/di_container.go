package di

import (
	"go.uber.org/dig"
)

type Graph struct {
	dig *dig.Container
}

func NewGraph() *Graph {
	return &Graph{dig: dig.New()}
}

func (g *Graph) Provide(constructor interface{}) error {
	return g.dig.Provide(constructor)
}

func (g *Graph) Invoke(function interface{}) error {
	return g.dig.Invoke(function)
}

func (g *Graph) String() string {
	return g.dig.String()
}

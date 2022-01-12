package dag_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/ctxmeta"
	"github.com/DoNewsCode/core/dag"
)

// This example shows how to pass results to the next vertex or the dag caller.
func ExampleDAG_run() {
	d := dag.New()
	v1 := d.AddVertex(func(ctx context.Context) error {
		ctxmeta.GetBaggage(ctx).Set("v1Result", "foo")
		return nil
	})
	v2 := d.AddVertex(func(ctx context.Context) error {
		v1Result, _ := ctxmeta.GetBaggage(ctx).Get("v1Result")
		fmt.Println(v1Result)
		return nil
	})
	d.AddEdge(v1, v2)

	_, ctx := ctxmeta.Inject(context.Background())
	d.Run(ctx)

	// Output:
	// foo
}

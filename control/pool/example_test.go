package pool_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/control/pool"

	"github.com/gorilla/mux"
)

func Example() {
	c := core.Default(
		core.WithInline("http.addr", ":9777"),
		core.WithInline("log.level", "none"),
	)
	c.Provide(pool.Providers())

	c.Invoke(func(m *pool.Manager, dispatcher lifecycle.HTTPServerStart) {
		p := pool.NewPool(m, 10)
		dispatcher.On(func(ctx context.Context, payload lifecycle.HTTPServerStartPayload) error {
			go func() {
				if _, err := http.Get("http://localhost:9777/"); err != nil {
					panic(err)
				}
			}()
			return nil
		})
		c.AddModule(core.HttpFunc(func(router *mux.Router) {
			router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
				p.Go(request.Context(), func(asyncContext context.Context) {
					select {
					case <-asyncContext.Done():
						fmt.Println("async context cancelled")
					case <-time.After(time.Second):
						fmt.Println("async context will not be cancelled")
					}
				})
				writer.Write(nil)
			})
		}))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c.Serve(ctx)

	// Output:
	// async context will not be cancelled

}

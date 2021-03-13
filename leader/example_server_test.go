// +build integration

package leader_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/leader"
	"github.com/DoNewsCode/core/otetcd"
	"github.com/gorilla/mux"
)

type ServerModule struct {
	Sts *leader.Status
}

func (s ServerModule) ProvideHTTP(router *mux.Router) {
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if s.Sts.IsLeader() {
			writer.Write([]byte("I am leader"))
		} else {
			writer.Write([]byte("I am follower"))
		}
	})
}

func Example_server() {
	c := core.Default(core.WithInline("log.level", "none"))
	c.Provide(otetcd.Providers())
	c.Provide(leader.Providers())
	c.Invoke(func(dispatcher contract.Dispatcher) {
		// This listener will be called twice. Once on becoming the leader and once on resigning the leader.
		dispatcher.Subscribe(events.Listen(events.From(&leader.Status{}), func(ctx context.Context, event contract.Event) error {
			fmt.Println(event.Data().(*leader.Status).IsLeader())
			return nil
		}))
	})
	c.Invoke(func(sts *leader.Status) {
		c.AddModule(ServerModule{Sts: sts})
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c.Serve(ctx)

	// Output:
	// true
	// false
}

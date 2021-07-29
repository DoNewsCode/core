package otes_test

import (
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otes"
	"github.com/olivere/elastic/v7"
)

func Example() {
	c := core.New(core.WithInline("log.level", "none"))
	c.ProvideEssentials()
	c.Provide(otes.Providers())
	c.Invoke(func(esClient *elastic.Client) {
		running := esClient.IsRunning()
		fmt.Println(running)
	})
	// Output:
	// true
}

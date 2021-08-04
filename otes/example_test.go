package otes_test

import (
	"fmt"
	"os"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otes"
	"github.com/olivere/elastic/v7"
)

func Example() {
	if os.Getenv("ELASTICSEARCH_ADDR") == "" {
		fmt.Println("set ELASTICSEARCH_ADDR to run example")
		return
	}
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

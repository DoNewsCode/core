package otes_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otes"
	"github.com/olivere/elastic/v7"
)

func Example_providers() {
	if os.Getenv("ELASTICSEARCH_ADDR") == "" {
		fmt.Println("set ELASTICSEARCH_ADDR to run example")
		return
	}
	c := core.New(
		core.WithInline("log.level", "none"),
		core.WithInline("es.default.url", strings.Split(os.Getenv("ELASTICSEARCH_ADDR"), ",")),
	)
	c.ProvideEssentials()
	c.Provide(otes.Providers(otes.WithClientConstructor(func(args otes.ClientConstructorArgs) (*elastic.Client, error) {
		return elastic.NewClient(elastic.SetURL(args.Conf.URL...), elastic.SetGzip(true))
	})))
	c.Invoke(func(esClient *elastic.Client) {
		running := esClient.IsRunning()
		fmt.Println(running)
	})
	// Output:
	// true
}

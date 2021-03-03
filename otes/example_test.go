package otes_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otes"
	"github.com/olivere/elastic/v7"
)

func Example() {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(otes.Providers())
	c.Invoke(func(esClient *elastic.Client) {
		_, _ ,err := esClient.Ping("http://localhost:9200").Do(context.TODO())
		fmt.Println(err)
	})
	// Output:
	// nil
}

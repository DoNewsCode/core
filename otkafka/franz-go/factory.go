package franz_go

import (
	"github.com/DoNewsCode/core/di"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Factory is a *di.Factory that creates *kafka.Client.
//
// Unlike other database providers, the kafka factories don't bundle a default
// kafka reader/writer. It is suggested to use Topic name as the identifier of
// kafka config rather than an opaque name such as default.
type Factory struct {
	*di.Factory
}

// Make returns a *kgo.Client under the provided configuration entry.
func (k Factory) Make(name string) (*kgo.Client, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kgo.Client), nil
}

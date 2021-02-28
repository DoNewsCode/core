/*
Package kitkafka provides a kafka transport for go kit.

Introduction

Go kit has some great properties, such as allowing multiple transport to be used
simultaneously. Sadly it limits itself to only support RPCs. in real projects
with many decoupled component, messaging is an inevitable path we must go down.

Go kit models the RPCs as:

	func(context.Context, request interface{}) (response interface{}, err error)

Package kitkafka treat messaging as a special case of RPC, where the response is
always ignored. By using the same model, package kitkafka brings all go kit
endpoint into the hood.

See examples for go kit project with kafka as transport.

Integration

kitkafka exports the configuration in this format:

	kafka:
	  writer:
		foo:
		  brokers:
			- localhost:9092
		  topic: foo
	  reader:
		bar:
		  brokers:
			- localhost:9092
		  topic: bar
		  groupID: bar-group

For a complete overview of all available options, call the config init command.

To use package kitkafka with package core, add:

	var c *core.C = core.New()
	c.Provide(kitkafka.provideKafkaFactory)

The reader and writer factories are bundled into that single provider.

Standalone Usage

in some scenarios, the whole go kit family might be overkill. To directly
interact with kafka, use the factory to make writers and readers. Those
writers/readers are provided by github.com/segmentio/kafka-go.

	c.Invoke(func(writer *kafka.Writer) {
		writer.WriteMessage(kafka.Message{})
	})

*/
package kitkafka

/*
Package otkafka contains the opentracing integrated a kafka transport for package Core.
The underlying kafka library is kafka-go: https://github.com/segmentio/kafka-go.

Integration

otkafka exports the configuration in this format:

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

To use package otkafka with package core, add:

	var c *core.C = core.New()
	c.Provide(otkafka.Providers())

The reader and writer factories are bundled into that single provider.

Standalone Usage

in some scenarios, the whole go kit family might be overkill. To directly
interact with kafka, use the factory to make writers and readers. Those
writers/readers are provided by github.com/segmentio/kafka-go.

	c.Invoke(func(writer *kafka.Writer) {
		writer.WriteMessage(kafka.Message{})
	})

*/
package otkafka

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/movio/kasper"
)

// ProducerExample is Kafka message processor that shows how to write messages to Kafka topics
type ProducerExample struct{}

// Process processes Kafka messages from topics "hello" and "world" and publish outgoing messages to "world" topi
func (*ProducerExample) Process(msg kasper.IncomingMessage, sender kasper.Sender, coordinator kasper.Coordinator) {
	key := msg.Key.(string)
	value := msg.Value.(string)
	offset := msg.Offset
	topic := msg.Topic
	partition := msg.Partition
	format := "Got message: key='%s', value='%s' at offset='%d' (topic='%s', partition='%d')\n"
	fmt.Printf(format, key, value, offset, topic, partition)
	outgoingMessage := kasper.OutgoingMessage{
		Topic:     "world",
		Partition: 0,
		Key:       msg.Key,
		Value:     fmt.Sprintf("Hello %s", msg.Value),
	}
	sender.Send(outgoingMessage)
}

func main() {
	config := kasper.TopicProcessorConfig{
		TopicProcessorName: "producer-example",
		BrokerList:         []string{"localhost:9092"},
		InputTopics:        []string{"hello", "world"},
		TopicSerdes: map[string]kasper.TopicSerde{
			"hello": {
				KeySerde:   kasper.NewStringSerde(),
				ValueSerde: kasper.NewStringSerde(),
			},
			"world": {
				KeySerde:   kasper.NewStringSerde(),
				ValueSerde: kasper.NewStringSerde(),
			},
		},
		ContainerCount: 1,
		PartitionToContainerID: map[int]int{
			0: 0,
		},
		AutoMarkOffsetsInterval: 100 * time.Millisecond,
		Config:                  kasper.DefaultConfig(),
	}
	mkMessageProcessor := func() kasper.MessageProcessor { return &ProducerExample{} }
	topicProcessor := kasper.NewTopicProcessor(&config, mkMessageProcessor, 0)
	topicProcessor.Start()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Topic processor is running...")
	for range signals {
		signal.Stop(signals)
		topicProcessor.Shutdown()
		break
	}
	log.Println("Topic processor shutdown complete.")
}

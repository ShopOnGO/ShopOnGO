package kafkaService

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type KafkaProducers map[string]*KafkaService

func InitKafkaProducers(brokers []string, topics map[string]string) KafkaProducers {
	producers := make(KafkaProducers)

	for name, topic := range topics {
		producers[name] = NewProducer(brokers, topic)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		for name, producer := range producers {
			fmt.Printf("🛑 Закрытие Kafka-писателя для %s\n", name)
			producer.Close()
		}
		os.Exit(0)
	}()

	return producers
}

package kafkaService

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShopOnGO/ShopOnGO/configs"
)

type KafkaProducers map[string]*KafkaService

func InitKafkaProducers(conf *configs.Config) KafkaProducers {
	producers := make(KafkaProducers)

	for name, topic := range conf.Kafka.Topics {
		producers[name] = NewProducer(conf.Kafka.Brokers, topic)
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

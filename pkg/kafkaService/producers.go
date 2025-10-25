package kafkaService

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type KafkaProducers map[string]*KafkaService

func InitKafkaProducers(brokers []string, topics map[string]string) KafkaProducers {
	producers := make(KafkaProducers)

	for name, topic := range topics {
		var producer *KafkaService

		// до 10 попыток подключения к Kafka
		for i := 1; i <= 10; i++ {
			producer = NewProducer(brokers, topic)
			if producer != nil {
				fmt.Printf("✅ Kafka producer для %s подключен (попытка %d)\n", name, i)
				break
			}
			fmt.Printf("⚠️ Kafka producer для %s не готов, попытка %d/10\n", name, i)
			time.Sleep(5 * time.Second)
		}

		if producer == nil {
			fmt.Printf("❌ Не удалось подключиться к Kafka для %s после 10 попыток\n", name)
			continue
		}

		producers[name] = producer
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

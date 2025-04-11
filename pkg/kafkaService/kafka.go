package kafkaService

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

// --- PRODUCER ---

func NewProducer(brokers []string, topic string) *KafkaService {
	waitForKafka(brokers, 30*time.Second)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireOne),
		BatchTimeout: 10 * time.Millisecond,
	})

	return &KafkaService{
		Writer: writer,
	}
}

// Produce отправляет сообщение в Kafka.
func (k *KafkaService) Produce(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	return k.Writer.WriteMessages(ctx, msg)
}


func waitForKafka(brokers []string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := kafka.Dial("tcp", brokers[0])
		if err == nil {
			defer conn.Close()
			_, err = conn.ApiVersions()
			if err == nil {
				fmt.Println("✅ Kafka доступна")
				return
			}
		}
		fmt.Printf("⏳ Ожидание Kafka (%s): %v\n", brokers[0], err)
		time.Sleep(2 * time.Second)
	}
	panic(fmt.Sprintf("❌ Kafka недоступна после %v", timeout))
}

func (k *KafkaService) Close() error {
	if k.Writer != nil {
		if err := k.Writer.Close(); err != nil {
			return err
		}
	}
	if k.Reader != nil {
		return k.Reader.Close()
	}
	return nil
}

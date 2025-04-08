package kafkaService

import (
	"context"
	"fmt"
	"strings"
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

	if err := ensureTopicInitialized(brokers, topic); err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second) // немного подождать для инициализации

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

func ensureTopicInitialized(brokers []string, topic string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	defer writer.Close()

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte("dummy"),
			Value: []byte("init"),
		})
		if err == nil {
			fmt.Printf("✅ Топик '%s' инициализирован\n", topic)
			return nil
		}
		if strings.Contains(err.Error(), "topic partition has no leader") {
			fmt.Println("Топик без лидера, повтор через 5 секунд...")
			time.Sleep(5 * time.Second)
			continue
		}
		return fmt.Errorf("не удалось отправить тестовое сообщение: %w", err)
	}
	return fmt.Errorf("тестовое сообщение не отправлено после %d попыток", maxRetries)
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

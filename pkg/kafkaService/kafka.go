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

// --- CONSUMER ---

func NewConsumer(brokers []string, topic, groupID, clientID string) *KafkaService {
	waitForKafka(brokers, 30*time.Second)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 1,   // получать сразу, даже маленькие
		MaxBytes: 1e6, // до 1MB в batch
		MaxWait:  10 * time.Millisecond,
		// MinBytes: 10e3,// Для большого количества уведомлений
		// MaxBytes: 10e6,
		// Dialer: &kafka.Dialer{
		// 	Timeout:  10 * time.Second,
		// 	ClientID: clientID,
		// },
	})

	return &KafkaService{
		Reader: reader,
	}
}

func (k *KafkaService) Consume(ctx context.Context, handler func(message kafka.Message) error) {
	for {
		msg, err := k.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Println("Контекст отменён, остановка консьюмера")
				break
			}
			if strings.Contains(err.Error(), "Group Coordinator Not Available") {
				fmt.Println("Координатор группы не доступен, повтор через 5 секунд...")
				time.Sleep(5 * time.Second)
				continue
			}
			fmt.Printf("Ошибка чтения: %v\n", err)
			continue
		}

		if err := handler(msg); err != nil {
			fmt.Printf("Ошибка обработки: %v\n", err)
		}
	}
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

func (k *KafkaService) Produce(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	return k.Writer.WriteMessages(ctx, msg)
}

// ProduceMessage позволяет передавать кастомный kafka.Message
func (k *KafkaService) ProduceMessage(ctx context.Context, msg kafka.Message) error {
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

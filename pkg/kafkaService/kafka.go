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

// NewKafkaService инициализирует продюсера и консьюмера.
// brokers — список адресов Kafka-брокеров,
// topic — название топика,
// groupID — идентификатор группы для консьюмера.
func NewKafkaService(brokers []string, topic, groupID string) *KafkaService {
	
	waitForKafka(brokers, 30*time.Second)

	// Инициализируем топик (тестовое продюсирование)
    if err := ensureTopicInitialized(brokers, topic); err != nil {
        panic(err)
    }
    // Даем Kafka немного времени на обновление метаданных
    time.Sleep(5 * time.Second)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireOne),
		BatchTimeout: 10 * time.Millisecond,
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
		Dialer: &kafka.Dialer{
			Timeout:  10 * time.Second,
			ClientID: "review-consumer",
		},
	})

	return &KafkaService{
		Writer: writer,
		Reader: reader,
	}
}

func waitForKafka(brokers []string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := kafka.Dial("tcp", brokers[0])
		if err == nil {
			defer conn.Close()
			_, err = conn.ApiVersions()
			if err == nil {
				fmt.Println("✅ Kafka готова к приёму API-запросов")
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
            fmt.Printf("✅ Топик '%s' успешно инициализирован\n", topic)
            return nil
        }
        // Если ошибка связана с отсутствием лидера для партиции, ждем и повторяем попытку
        if strings.Contains(err.Error(), "topic partition has no leader") {
            fmt.Println("Топик существует, но лидер ещё не назначен, повторная попытка через 5 секунд...")
            time.Sleep(5 * time.Second)
            continue
        }
        return fmt.Errorf("не удалось продюсировать тестовое сообщение: %w", err)
    }
    return fmt.Errorf("не удалось продюсировать тестовое сообщение после %d попыток", maxRetries)
}


// Produce отправляет сообщение в Kafka.
func (k *KafkaService) Produce(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	return k.Writer.WriteMessages(ctx, msg)
}

// Consume запускает бесконечный цикл чтения сообщений.
// Для каждого полученного сообщения вызывается переданный обработчик.
func (k *KafkaService) Consume(ctx context.Context, handler func(message kafka.Message) error) {
	for {
		msg, err := k.Reader.ReadMessage(ctx)
		if err != nil {
			// Если контекст отменен, выходим из цикла
			if ctx.Err() != nil {
				fmt.Println("Контекст отменен, остановка консьюмера")
				break
			}

			// Если ошибка связана с недоступностью координатора группы, ждем и повторяем попытку
			if strings.Contains(err.Error(), "Group Coordinator Not Available") {
				fmt.Println("Координатор группы не доступен, повторная попытка через 5 секунд...")
				time.Sleep(5 * time.Second)
				continue
			}

			fmt.Printf("Ошибка при чтении сообщения: %v\n", err)
			continue
		}

		if err := handler(msg); err != nil {
			fmt.Printf("Ошибка обработки сообщения: %v\n", err)
		}
	}
}

// Close закрывает подключения продюсера и консьюмера.
func (k *KafkaService) Close() error {
	if err := k.Writer.Close(); err != nil {
		return err
	}
	return k.Reader.Close()
}

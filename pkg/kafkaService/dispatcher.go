package kafkaService

import (
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Dispatcher хранит мапу ключей → обработчики
type Dispatcher struct {
	handlers map[string]func(msg kafka.Message) error
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{handlers: make(map[string]func(msg kafka.Message) error)}
}

func (d *Dispatcher) Register(key string, handler func(msg kafka.Message) error) {
	d.handlers[key] = handler
}

func (d *Dispatcher) Dispatch(msg kafka.Message) error {
	key := string(msg.Key)

	// Итерируемся по всем зарегистрированным префиксам
	for prefix, handler := range d.handlers {
		// Если ключ сообщения начинается с префикса
		if strings.HasPrefix(key, prefix) {
			return handler(msg) // Вызываем нужный обработчик
		}
	}

	// Если ни один префикс не подошел
	return fmt.Errorf("no handler for key %q", key)
}

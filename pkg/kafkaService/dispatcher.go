package kafkaService

import (
	"fmt"

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
	if handler, ok := d.handlers[key]; ok {
		return handler(msg)
	}
	return fmt.Errorf("no handler for key %q", key)
}

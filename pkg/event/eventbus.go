package event

const ( // описание всех событий:
	LInkVisitedEvent = "link.visited"
)

type Event struct {
	Type string
	Data any
}

type EventBus struct {
	bus chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		bus: make(chan Event),
	}
}

func (e *EventBus) Publish(event Event) { // метод публикации(отправитель)
	e.bus <- event
}

func (e *EventBus) Subscribe() <-chan Event { // получение СООБЩЕНИЯ!! канала  <-chan!!
	return e.bus
}

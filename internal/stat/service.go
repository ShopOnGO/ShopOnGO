package stat

import (
	"log"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/event"
)

type StatServiceDeps struct {
	EventBus       *event.EventBus
	StatRepository *StatRepository
}

type StatService struct {
	EventBus       *event.EventBus
	StatRepository *StatRepository
}

func NewStatService(deps *StatServiceDeps) *StatService {
	return &StatService{
		EventBus:       deps.EventBus,
		StatRepository: deps.StatRepository,
	}
}

func (s *StatService) AddClick() {
	for msg := range s.EventBus.Subscribe() { // получаем список наших сообщений
		if msg.Type == event.LInkVisitedEvent {
			linkId, ok := msg.Data.(uint)
			if !ok {
				log.Fatalln("Bad LInkVisitedEvent Data", msg.Data)
				continue
			}
			s.StatRepository.AddClick(linkId)
		}
	}
}

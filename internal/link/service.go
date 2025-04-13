package link

import (
	"errors"
	"math/rand"

	"github.com/ShopOnGO/ShopOnGO/pkg/event"
	"gorm.io/gorm"
)

type LinkService struct {
	LinkRepository *LinkRepository
	EventBus       *event.EventBus
}

func NewLinkService(linkRepository *LinkRepository, eventBus *event.EventBus) *LinkService {
	return &LinkService{
		LinkRepository: linkRepository,
		EventBus:       eventBus,
	}
}

func (s *LinkService) CreateLink(url string) (*Link, error) {
	link := NewLink(url)
	for {
		existedLink, _ := s.LinkRepository.GetByHash(link.Hash)
		if existedLink == nil {
			break
		}
		link.GenerateHash()
	}
	return s.LinkRepository.Create(link)
}

func (s *LinkService) UpdateLink(id uint, url, hash string) (*Link, error) {
	if id == 0 {
		return nil, errors.New("invalid link ID")
	}

	link := &Link{
		Model: gorm.Model{ID: id},
		Url:   url,
		Hash:  hash,
	}

	return s.LinkRepository.Update(link)
}

func (s *LinkService) DeleteLink(id uint) error {
	if id == 0 {
		return errors.New("invalid link ID")
	}

	_, err := s.LinkRepository.GetById(id)
	if err != nil {
		return errors.New("link not found")
	}

	return s.LinkRepository.Delete(id)
}

func (s *LinkService) GoTo(hash string) (*Link, error) {
	link, err := s.LinkRepository.GetByHash(hash)
	if err != nil {
		return nil, err
	}

	// Публикуем событие о посещении ссылки
	go s.EventBus.Publish(event.Event{
		Type: event.LInkVisitedEvent,
		Data: link.ID,
	})

	return link, nil
}

func (s *LinkService) GetLinkByHash(hash string) (*Link, error) {
	return s.LinkRepository.GetByHash(hash)
}

func (s *LinkService) GetLinkByID(id uint) (*Link, error) {
	if id == 0 {
		return nil, errors.New("invalid link ID")
	}
	return s.LinkRepository.GetById(id)
}

func (s *LinkService) GetAll(limit, offset int) ([]Link, int64, error) {
	links := s.LinkRepository.GetAll(limit, offset)
	count := s.LinkRepository.Count()
	return links, count, nil
}

func (s *LinkService) CountLinks() int64 {
	return s.LinkRepository.Count()
}

func NewLink(url string) *Link {
	link := &Link{
		Url: url,
	}
	link.GenerateHash()
	return link
}

func (link *Link) GenerateHash() {
	link.Hash = RandStringRunes(10)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

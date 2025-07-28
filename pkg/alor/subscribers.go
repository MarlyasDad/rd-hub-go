package alor

import (
	"errors"
	"sync"
)

func NewSubscribers() Subscribers {
	return Subscribers{
		toAdd:    make(map[SubscriberID]*Subscriber),
		toDelete: make([]SubscriberID, 0),
		list:     make(map[SubscriberID]*Subscriber),
	}
}

type Subscribers struct {
	toAdd    map[SubscriberID]*Subscriber
	toDelete []SubscriberID
	list     map[SubscriberID]*Subscriber
	mu       sync.RWMutex
}

func (s *Subscribers) Add(subscriber *Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.toAdd[subscriber.ID] = subscriber
}

func (s *Subscribers) Delete(subscriberID SubscriberID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.toDelete = append(s.toDelete, subscriberID)
}

func (s *Subscribers) Get(subscriberID SubscriberID) (*Subscriber, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subscriber, ok := s.list[subscriberID]
	if !ok {
		return nil, errors.New("subscriber not found")
	}

	return subscriber, nil
}

func (s *Subscribers) All() map[SubscriberID]*Subscriber {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.list
}

func (s *Subscribers) Rebalancing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Добавляем подписчиков
	for k, v := range s.toAdd {
		// если есть существующий, то перезаписываем,
		// но так быть не должно, так как уникальные ID
		s.list[k] = v
	}

	// удаляем подписчиков
	for _, key := range s.toDelete {
		delete(s.list, key)
	}
}

package alor

import (
	"errors"
	"sync"
)

func NewSubscriptions() Subscriptions {
	return Subscriptions{
		toAdd:    make(map[GUID]SubscriptionContainer),
		toDelete: make(map[GUID]SubscriberID),
		list:     make(map[GUID]SubscriptionContainer),
	}
}

type (
	GUID string

	SubscriptionContainer struct {
		Subscription *Subscription
		Active       bool
		Items        map[SubscriberID]bool
		//Subscriber SubscriberID
		// Несколько подписок - один подписчик (1кМ)
		// Вместо контейнера всё в подписке?
	}

	Subscriptions struct {
		toAdd    map[GUID]SubscriptionContainer
		toDelete map[GUID]SubscriberID
		list     map[GUID]SubscriptionContainer
		active   bool
		mu       sync.RWMutex
	}
)

func (s *Subscriptions) Add(subscriberID SubscriberID, subscription *Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	container := SubscriptionContainer{
		Subscription: subscription,
		Active:       false,
		Items: map[SubscriberID]bool{
			subscriberID: true,
		},
	}

	s.toAdd[GUID(subscription.GUID)] = container
}

func (s *Subscriptions) SetActive(guid GUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriptionContainer := s.list[guid]
	subscriptionContainer.Active = true
	s.list[guid] = subscriptionContainer
}

func (s *Subscriptions) Delete(subscriberID SubscriberID, guid GUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Получаем контейнер подписки
	_, ok := s.list[guid]
	if !ok {
		return errors.New("the subscription is not exists")
	}

	// Удаляем подписчика из подписки
	delete(s.list[guid].Items, subscriberID)

	// Удаляем подписку если она пустая
	if len(s.list[guid].Items) == 0 {
		delete(s.list, guid)
	}

	return nil
}

func (s *Subscriptions) Get(guid GUID) (SubscriptionContainer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subscriptionContainer, ok := s.list[guid]
	if !ok {
		return subscriptionContainer, errors.New("subscription container not found")
	}

	return subscriptionContainer, nil
}

func (s *Subscriptions) All() (map[GUID]SubscriptionContainer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.list, nil
}

func (s *Subscriptions) Rebalancing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Добавляем подписки
	for guid, container := range s.toAdd {
		// Создать подписку если она не существует
		_, ok := s.list[guid]
		if !ok {
			// Создаём подписку заново
			s.list[guid] = container
		} else {
			// Копируем подписчиков в уже существующую
			for subscriberID, _ := range container.Items {
				s.list[guid].Items[subscriberID] = true
			}
		}
	}

	// Удаляем подписки
	for guid, subscriberID := range s.toDelete {
		delete(s.list[guid].Items, subscriberID)

		// если подписчиков нет,
		if len(s.list[guid].Items) == 0 {
			delete(s.list, guid)
		}
	}
}

package subscribers

import (
	"errors"
	"github.com/MarlyasDad/rd-hub-go/internal/domain"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
)

func (s Service) GetSubscriberBars(subscriberID alor.SubscriberID, heikenAshi bool) ([]*alor.Bar, error) {
	bars, err := s.brokerClient.GetAllSubscriberBars(subscriberID)
	if err != nil {
		if errors.Is(err, alor.ErrSubscriberNotFound) {
			return nil, domain.ErrSubscriberNotFound
		}
	}

	return bars, nil
}

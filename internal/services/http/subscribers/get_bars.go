package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
)

func (s Service) GetSubscriberBars(subscriberID uuid.UUID, heikenAshi bool) ([]*alor.Bar, error) {
	return s.brokerClient.GetAllSubscriberBars(subscriberID)
}

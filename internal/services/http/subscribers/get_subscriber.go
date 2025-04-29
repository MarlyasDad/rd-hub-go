package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
)

func (s Service) GetSubscriber(subscriberID uuid.UUID) (*alor.Subscriber, error) {
	return s.brokerClient.GetSubscriber(subscriberID)
}

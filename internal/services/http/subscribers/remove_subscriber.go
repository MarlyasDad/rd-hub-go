package subscribers

import (
	"github.com/google/uuid"
)

func (s Service) RemoveSubscriber(subscriberID uuid.UUID) error { // FromAlltrades
	return s.brokerClient.RemoveSubscriber(subscriberID)
}

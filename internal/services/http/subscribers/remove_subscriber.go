package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
)

func (s Service) RemoveSubscriber(subscriberID alor.SubscriberID) error { // FromAlltrades
	return s.brokerClient.RemoveSubscriber(subscriberID)
}

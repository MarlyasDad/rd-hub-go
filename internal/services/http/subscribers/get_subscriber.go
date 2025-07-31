package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
)

func (s Service) GetSubscriber(subscriberID alor.SubscriberID) (*alor.Subscriber, error) {
	return s.brokerClient.GetSubscriber(subscriberID)
}

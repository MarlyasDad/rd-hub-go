package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
)

type brokerClient interface {
	GetSubscribers() []*alor.Subscriber
	GetAllSubscriberBars(subscriberID alor.SubscriberID) ([]*alor.Bar, error)
	AddSubscriber(subscriber *alor.Subscriber) error
	RemoveSubscriber(subscriberID alor.SubscriberID) error
	GetAllTrades(params alor.GetAllTradesV2Params) ([]alor.AllTradesSlimData, error)
	GetSubscriber(subscriberID alor.SubscriberID) (*alor.Subscriber, error)
}

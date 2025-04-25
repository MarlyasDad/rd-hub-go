package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
)

type brokerClient interface {
	GetSubscribers() []*alor.Subscriber
	GetAllSubscriberBars(subscriberID uuid.UUID) ([]*alor.Bar, error)
	AddSubscriber(subscriber *alor.Subscriber) error
	RemoveSubscriber(subscriberID uuid.UUID) error
	GetAllTrades(params alor.GetAllTradesV2Params) ([]alor.AllTradesSlimData, error)
}

package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
)

func (s Service) GetSubscribers() []*alor.Subscriber {
	return s.brokerClient.GetSubscribers()
}

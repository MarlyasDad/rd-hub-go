package subscribers

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"net/http"

	httpSubscribersCommand "github.com/MarlyasDad/rd-hub-go/internal/services/http/subscribers"
)

func RegisterRoutes(mux *http.ServeMux, brokerClient *alor.Client) {
	getSubscribersPattern := "GET /api/subscriber/all"
	mux.Handle(
		getSubscribersPattern,
		NewSubscribersListHandler(
			httpSubscribersCommand.New(brokerClient),
			getSubscribersPattern,
		),
	)

	getSubscriberBarsPattern := "GET /api/subscriber/{subscriber_id}/bars"
	mux.Handle(
		getSubscriberBarsPattern,
		NewSubscriberBarsHandler(
			httpSubscribersCommand.New(brokerClient),
			getSubscriberBarsPattern,
		),
	)

	addSubscriberPattern := "POST /api/subscriber"
	mux.Handle(
		addSubscriberPattern,
		NewSubscriberAddHandler(
			httpSubscribersCommand.New(brokerClient),
			addSubscriberPattern,
		),
	)
}

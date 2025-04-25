package subscribers

import (
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/google/uuid"
	"net/http"
)

type (
	subscriberAddCommand interface {
		AddSubscriber() (uuid.UUID, error)
	}

	SubscriberAddHandler struct {
		name                 string
		subscriberAddCommand subscriberAddCommand
	}

	//GetSubscribersListResponse struct {
	//	ID          uuid.UUID `json:"id"`
	//	Description string    `json:"description"`
	//	CreatedAt   time.Time `json:"createdAt"`
	//	ClientID    int64     `json:"clientID"`
	//	Exchange    string    `json:"exchange"`
	//	Code        string    `json:"code"`
	//	Board       string    `json:"board"`
	//	Timeframe   int64     `json:"timeframe"`
	//}
)

func NewSubscriberAddHandler(command subscriberAddCommand, name string) *SubscriberAddHandler {
	return &SubscriberAddHandler{
		name:                 name,
		subscriberAddCommand: command,
	}
}

func (h *SubscriberAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ctx = r.Context()
		err error
	)

	subscriberID, err := h.subscriberAddCommand.AddSubscriber()
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
		return
	}

	bodyBytes, _ := json.Marshal(subscriberID)

	responses.GetSuccessResponseWithBody(w, bodyBytes)
}

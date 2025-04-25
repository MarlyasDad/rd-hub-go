package subscribers

import (
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
)

type (
	subscriberRemoveCommand interface {
		RemoveSubscriber(uuid.UUID) error
	}

	SubscriberRemoveHandler struct {
		name                    string
		subscriberRemoveCommand subscriberRemoveCommand
	}

	subscriberRemoveRequest struct {
		ID uuid.UUID `json:"id"`
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

func NewSubscriberRemoveHandler(command subscriberAddCommand, name string) *SubscriberAddHandler {
	return &SubscriberAddHandler{
		name:                 name,
		subscriberAddCommand: command,
	}
}

func (h *SubscriberRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ctx = r.Context()
		requestData *subscriberRemoveRequest
		err         error
	)

	if requestData, err = h.getRequestData(r); err != nil {
		// Неправильный формат запроса
		responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	if err = h.validateRequestData(requestData); err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	err = h.subscriberRemoveCommand.RemoveSubscriber(requestData.ID)
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
		return
	}

	// bodyBytes, _ := json.Marshal(subscriberID)

	responses.GetSuccessResponse(w)
}

func (h *SubscriberRemoveHandler) getRequestData(r *http.Request) (requestData *subscriberRemoveRequest, err error) {
	requestData = &subscriberRemoveRequest{}

	if err = json.NewDecoder(r.Body).Decode(requestData); err != nil {
		return
	}

	return
}

func (h *SubscriberRemoveHandler) validateRequestData(requestData *subscriberRemoveRequest) error {
	return validator.New().Struct(requestData)
}

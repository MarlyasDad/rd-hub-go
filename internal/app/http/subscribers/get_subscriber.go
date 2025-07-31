package subscribers

import (
	"encoding/json"
	"fmt"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
	"net/http"
)

type (
	getSubscriberCommand interface {
		GetSubscriber(subscriberID alor.SubscriberID) (*alor.Subscriber, error)
	}

	GetSubscriberHandler struct {
		name                 string
		getSubscriberCommand getSubscriberCommand
	}

	GetSubscriberRequest struct {
		SubscriberID alor.SubscriberID `json:"subscriber_id"`
	}
)

func NewSubscriberHandler(command getSubscriberCommand, name string) *GetSubscriberHandler {
	return &GetSubscriberHandler{
		name:                 name,
		getSubscriberCommand: command,
	}
}

func (h *GetSubscriberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ctx = r.Context()
		requestData *GetSubscriberRequest
		err         error
	)

	if requestData, err = h.getRequestData(r); err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	subscribers, err := h.getSubscriberCommand.GetSubscriber(requestData.SubscriberID)
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
	}

	subscribersJson, err := json.Marshal(subscribers)
	if err != nil {
		responses.GetErrorResponse(w, h.name, fmt.Errorf("json marshalling failed: %w", err), http.StatusInternalServerError)
		return
	}

	responses.GetSuccessResponse(w, subscribersJson)
}

func (h *GetSubscriberHandler) getRequestData(r *http.Request) (requestData *GetSubscriberRequest, err error) {
	requestData = &GetSubscriberRequest{}

	subscriberID, err := uuid.Parse(r.PathValue("subscriber_id"))
	if err != nil {
		return
	}

	requestData.SubscriberID = alor.SubscriberID(subscriberID)

	return
}

//func (h *GetSubscriberHandler) validateRequestData(requestData *subtaskCreateRequest) error {
//	return validator.New().Struct(requestData)
//}

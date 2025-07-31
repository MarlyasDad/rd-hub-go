package subscribers

import (
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type (
	removeSubscriberCommand interface {
		RemoveSubscriber(id alor.SubscriberID) error
	}

	RemoveSubscriberHandler struct {
		name                    string
		removeSubscriberCommand removeSubscriberCommand
	}

	removeSubscriberRequest struct {
		ID alor.SubscriberID `json:"id"`
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

func NewRemoveSubscriberHandler(command removeSubscriberCommand, name string) *RemoveSubscriberHandler {
	return &RemoveSubscriberHandler{
		name:                    name,
		removeSubscriberCommand: command,
	}
}

func (h *RemoveSubscriberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ctx = r.Context()
		requestData *removeSubscriberRequest
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

	err = h.removeSubscriberCommand.RemoveSubscriber(requestData.ID)
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
		return
	}

	responses.GetSuccessResponse(w, nil)
}

func (h *RemoveSubscriberHandler) getRequestData(r *http.Request) (requestData *removeSubscriberRequest, err error) {
	requestData = &removeSubscriberRequest{}

	if err = json.NewDecoder(r.Body).Decode(requestData); err != nil {
		return
	}

	return
}

func (h *RemoveSubscriberHandler) validateRequestData(requestData *removeSubscriberRequest) error {
	return validator.New().Struct(requestData)
}

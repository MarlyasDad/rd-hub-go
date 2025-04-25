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
	getSubscriberBarsCommand interface {
		GetSubscriberBars(subscriberID uuid.UUID, heikenAshi bool) ([]*alor.Bar, error)
	}

	GetSubscriberBarsHandler struct {
		name                     string
		getSubscriberBarsCommand getSubscriberBarsCommand
	}

	GetSubscriberBarsRequest struct {
		SubscriberID uuid.UUID `json:"subscriber_id"`
		HeikenAshi   bool      `json:"heiken_ashi"`
	}
	GetSubscriberBarsResponse struct{}
)

func NewSubscriberBarsHandler(command getSubscriberBarsCommand, name string) *GetSubscriberBarsHandler {
	return &GetSubscriberBarsHandler{
		name:                     name,
		getSubscriberBarsCommand: command,
	}
}

func (h *GetSubscriberBarsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ctx = r.Context()
		requestData *GetSubscriberBarsRequest
		err         error
	)

	if requestData, err = h.getRequestData(r); err != nil {
		// Неправильный формат запроса
		responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	//if err = h.validateRequestData(requestData); err != nil {
	//	responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
	//	return
	//}

	subscribers, err := h.getSubscriberBarsCommand.GetSubscriberBars(requestData.SubscriberID, requestData.HeikenAshi)
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
	}

	portfoliosJson, err := json.Marshal(subscribers)
	if err != nil {
		responses.GetErrorResponse(w, h.name, fmt.Errorf("json marshalling failed: %w", err), http.StatusInternalServerError)
		return
	}

	responses.GetSuccessResponseWithBody(w, portfoliosJson)
}

func (h *GetSubscriberBarsHandler) getRequestData(r *http.Request) (requestData *GetSubscriberBarsRequest, err error) {
	requestData = &GetSubscriberBarsRequest{}

	subscriberID, err := uuid.Parse(r.PathValue("subscriber_id"))
	if err != nil {
		return
	}

	requestData.SubscriberID = subscriberID

	//heikenAshi, err := strconv.ParseBool(r.FormValue("heiken_ashi"))
	//if err == nil {
	//	requestData.HeikenAshi = heikenAshi
	//}

	return
}

//func (h *GetSubscriberBarsHandler) validateRequestData(requestData *subtaskCreateRequest) error {
//	return validator.New().Struct(requestData)
//}

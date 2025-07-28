package subscribers

import (
	"context"
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/MarlyasDad/rd-hub-go/internal/services/http/subscribers"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
)

type (
	addSubscriberCommand interface {
		AddSubscriber(ctx context.Context, params subscribers.AddSubscriberParams) (uuid.UUID, error)
	}

	AddSubscriberHandler struct {
		name                 string
		addSubscriberCommand addSubscriberCommand
	}

	addSubscriberRequest struct {
		Description   string               `json:"description" validate:"required"`
		Exchange      string               `json:"exchange" validate:"oneof=MOEX SPBX"`
		Code          string               `json:"code" validate:"required"`
		Board         string               `json:"board" validate:"required"`
		Timeframe     int64                `json:"timeframe"`
		Subscriptions requestSubscriptions `json:"subscriptions"`
		From          int64                `json:"from"` // Запрос истории с этой даты
	}

	requestSubscriptions struct {
		AllTrades requestAllTrades `json:"allTrades"`
		OrderBook requestOrderBook `json:"orderBook"`
		Bars      requestBars      `json:"bars"`
	}

	requestAllTrades struct {
		WithDelta            bool  `json:"withDelta"`
		WithMarketProfile    bool  `json:"withMarketProfile"`
		Depth                int64 `json:"depth"` // для захлёста истории во избежание потери данных
		IncludeVirtualTrades bool  `json:"includeVirtualTrades"`
	}

	requestOrderBook struct {
		WithOrderFlow bool  `json:"withOrderFlow"`
		Depth         int64 `json:"depth"` // глубина стакана в одну сторону
	}

	requestBars struct {
		WithOrderFlow bool  `json:"withOrderFlow"`
		Depth         int64 `json:"depth"` // глубина стакана в одну сторону
	}
)

func NewAddSubscriberHandler(command addSubscriberCommand, name string) *AddSubscriberHandler {
	return &AddSubscriberHandler{
		name:                 name,
		addSubscriberCommand: command,
	}
}

func (h *AddSubscriberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx         = r.Context()
		requestData *addSubscriberRequest
		err         error
	)

	if requestData, err = h.getRequestData(r); err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	// Унифицировать параметры, которые отвечают за подписки
	// если есть дельта, то обязательно должна быть подписка на обезличенные сделки
	// если есть стакан -> обязательная подписка на стакан
	// Сделать сабскрибера интерфейсом
	// Проверить подписки, если есть профиль или дельта, удалить подписку на свечи если существует
	params := subscribers.AddSubscriberParams{
		Description:       requestData.Description,
		Exchange:          requestData.Exchange,
		Code:              requestData.Code,
		Board:             requestData.Board,
		Timeframe:         requestData.Timeframe,
		WithDelta:         true, // сразу подписаться если не подписан, указать параметры
		WithMarketProfile: true, // сразу подписаться если не подписан, указать параметры
		WithOrderFlow:     true, // сразу подписаться если не подписан, указать параметры
	}

	subscriberID, err := h.addSubscriberCommand.AddSubscriber(ctx, params)
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
		return
	}

	bodyBytes, _ := json.Marshal(subscriberID)

	responses.GetSuccessResponseWithBody(w, bodyBytes)
}

func (h *AddSubscriberHandler) getRequestData(r *http.Request) (requestData *addSubscriberRequest, err error) {
	requestData = &addSubscriberRequest{}

	err = json.NewDecoder(r.Body).Decode(requestData)

	return
}

func (h *AddSubscriberHandler) validateRequestData(requestData *addSubscriberRequest) error {
	return validator.New().Struct(requestData)

	// write your custom validating logic here
}

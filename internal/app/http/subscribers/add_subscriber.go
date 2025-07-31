package subscribers

import (
	"context"
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/MarlyasDad/rd-hub-go/internal/services/http/subscribers"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type (
	addSubscriberCommand interface {
		AddSubscriber(ctx context.Context, params *subscribers.AddSubscriberParams) (alor.SubscriberID, error)
	}

	AddSubscriberHandler struct {
		name                 string
		addSubscriberCommand addSubscriberCommand
	}

	addSubscriberRequest struct {
		Description   string        `json:"description"`
		Instrument    Instrument    `json:"instrument"`
		Strategy      Strategy      `json:"strategy"`
		Subscriptions Subscriptions `json:"subscriptions"`
		Indicators    []Indicator   `json:"indicators"`
		Async         bool          `json:"async"`
	}

	Instrument struct {
		Exchange  string `json:"exchange"`
		Code      string `json:"code"`
		Board     string `json:"board"`
		Timeframe int64  `json:"timeframe"`
	}

	Strategy struct {
		Name                 string          `json:"name"`
		Settings             json.RawMessage `json:"settings"`
		WithDelta            bool            `json:"withDelta"`
		WithMarketProfile    bool            `json:"withMarketProfile"`
		WithOrderBookProfile bool            `json:"withOrderBookProfile"`
	}

	Subscriptions struct {
		AllTrades *AllTradesParams `json:"allTrades"`
		OrderBook *OrderBookParams `json:"orderBook"`
		Bars      *BarsParams      `json:"bars"`
	}

	AllTradesParams struct {
		Frequency            int  `json:"frequency"`
		Depth                int  `json:"depth"`
		IncludeVirtualTrades bool `json:"includeVirtualTrades"`
	}

	OrderBookParams struct {
		Frequency int `json:"frequency"`
		Depth     int `json:"depth"`
	}

	BarsParams struct {
		Frequency   int   `json:"frequency"`
		From        int64 `json:"from"`
		SkipHistory bool  `json:"skipHistory"`
		SplitAdjust bool  `json:"splitAdjust"`
	}

	Indicator struct {
		Name     string          `json:"name"`
		Settings json.RawMessage `json:"settings"`
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
		requestData *subscribers.AddSubscriberParams
		err         error
	)

	if requestData, err = h.getRequestData(r); err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	// добавляем подписчика

	subscriberID, err := h.addSubscriberCommand.AddSubscriber(ctx, requestData)
	if err != nil {
		responses.GetErrorResponse(w, h.name, err, http.StatusInternalServerError)
		return
	}

	bodyBytes, _ := json.Marshal(subscriberID)

	responses.GetSuccessResponse(w, bodyBytes)
}

func (h *AddSubscriberHandler) getRequestData(r *http.Request) (requestData *subscribers.AddSubscriberParams, err error) {
	requestData = &subscribers.AddSubscriberParams{}

	err = json.NewDecoder(r.Body).Decode(requestData)

	return
}

func (h *AddSubscriberHandler) validateRequestData(requestData *subscribers.AddSubscriberParams) error {
	return validator.New().Struct(requestData)

	// write your custom validating logic here
}

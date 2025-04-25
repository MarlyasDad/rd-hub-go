package subscribers

import (
	"encoding/json"
	"fmt"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/responses"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"net/http"
)

type (
	getSubscribersListCommand interface {
		GetSubscribers() []*alor.Subscriber
	}

	GetSubscribersListHandler struct {
		name                      string
		getSubscribersListCommand getSubscribersListCommand
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

func NewSubscribersListHandler(command getSubscribersListCommand, name string) *GetSubscribersListHandler {
	return &GetSubscribersListHandler{
		name:                      name,
		getSubscribersListCommand: command,
	}
}

func (h *GetSubscribersListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		// ctx = r.Context()
		err error
	)

	subscribers := h.getSubscribersListCommand.GetSubscribers()

	subscribersJson, err := json.Marshal(subscribers)
	if err != nil {
		responses.GetErrorResponse(w, h.name, fmt.Errorf("json marshalling failed: %w", err), http.StatusInternalServerError)
		return
	}

	responses.GetSuccessResponseWithBody(w, subscribersJson)
}

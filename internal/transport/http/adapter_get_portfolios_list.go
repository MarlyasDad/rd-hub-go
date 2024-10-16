package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	getPortfoliosListCommand interface {
		GetPortfoliosList(ctx context.Context) ([]string, error)
	}

	GetPortfoliosListHandler struct {
		name                     string
		getPortfoliosListCommand getPortfoliosListCommand
	}

	//getPortfoliosListRequest struct {
	//	// url params
	//	TaskId uuid.UUID `validate:"uuid,required"`
	//}
)

func NewPortfoliosListHandler(command getPortfoliosListCommand, name string) *GetPortfoliosListHandler {
	return &GetPortfoliosListHandler{
		name:                     name,
		getPortfoliosListCommand: command,
	}
}

func (h *GetPortfoliosListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		// requestData *getPortfoliosListRequest
		err error
	)

	// if requestData, err = h.getRequestData(r); err != nil {
	// 	GetErrorResponse(w, h.name, err, http.StatusBadRequest)
	// 	return
	// }

	// if err = h.validateRequestData(requestData); err != nil {
	// 	GetErrorResponse(w, h.name, err, http.StatusBadRequest)
	// 	return
	// }

	portfoliosList, err := h.getPortfoliosListCommand.GetPortfoliosList(ctx)

	if err != nil {
		GetErrorResponse(w, h.name, fmt.Errorf("command handler failed: %w", err), http.StatusInternalServerError)
		return
	}

	taskJson, err := json.Marshal(portfoliosList)
	if err != nil {
		GetErrorResponse(w, h.name, fmt.Errorf("json marshalling failed: %w", err), http.StatusInternalServerError)
		return
	}

	GetSuccessResponseWithBody(w, taskJson)
}

//func (h *GetTaskHandler) getRequestData(r *http.Request) (requestData *getTaskRequest, err error) {
//	requestData = &getTaskRequest{}
//
//	taskId, err := uuid.Parse(r.PathValue("task_id"))
//	if err != nil {
//		return
//	}
//
//	requestData.TaskId = taskId
//
//	return
//}

//func (h *GetTaskHandler) validateRequestData(requestData *getTaskRequest) error {
//	return validator.New().Struct(requestData)
//}

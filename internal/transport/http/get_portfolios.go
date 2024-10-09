package http

package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	taskCommand "vp_backend/internal/domain/websocket/task"
	"vp_backend/internal/entities"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type (
	getTaskCommand interface {
		GetTask(ctx context.Context, taskID uuid.UUID) (*entities.Task, error)
	}

	GetTaskHandler struct {
		name           string
		getTaskCommand getTaskCommand
	}

	getTaskRequest struct {
		// url params
		TaskId uuid.UUID `validate:"uuid,required"`
	}

	// {
	// 	"task_id": "string uuid",
	// 	"created_at": "2024-05-29T12:00:23.437Z",
	// 	"updated_at": "string",
	// 	"deleted_at?": "string | null",
	// 	"module": "enum [ddu_extraction, passport_extraction, doc_classification]",
	// 	"status": "enum [new, ready, error, changed, verified]",
	// 	"log?": {
	// 	  "response": "enum [successful, error]",
	// 	  "message?": "string"
	// 	},
	// 	"data?": "TaskData | null"
	// }

	// {
	//   task_id: number
	//   created_at: string
	//   updated_at: string
	//   result?: IDDUExtractionResult
	// }
)

func NewGetTaskHandler(command getTaskCommand, name string) *GetTaskHandler {
	return &GetTaskHandler{
		name:           name,
		getTaskCommand: command,
	}
}

func (h *GetTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx         = r.Context()
		requestData *getTaskRequest
		err         error
	)

	if requestData, err = h.getRequestData(r); err != nil {
		GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	if err = h.validateRequestData(requestData); err != nil {
		GetErrorResponse(w, h.name, err, http.StatusBadRequest)
		return
	}

	dataTask, err := h.getTaskCommand.GetTask(ctx, requestData.TaskId)

	if err != nil {
		if errors.Is(err, taskCommand.ErrNoContent) {
			GetErrorResponse(w, h.name, taskCommand.ErrNoContent, http.StatusNoContent)
			return
		}

		GetErrorResponse(w, h.name, fmt.Errorf("command handler failed: %w", err), http.StatusInternalServerError)
		return
	}

	taskJson, err := json.Marshal(dataTask)
	if err != nil {
		GetErrorResponse(w, h.name, fmt.Errorf("json marshalling failed: %w", err), http.StatusInternalServerError)
		return
	}

	GetSuccessResponseWithBody(w, taskJson)
}

func (h *GetTaskHandler) getRequestData(r *http.Request) (requestData *getTaskRequest, err error) {
	requestData = &getTaskRequest{}

	taskId, err := uuid.Parse(r.PathValue("task_id"))
	if err != nil {
		return
	}

	requestData.TaskId = taskId

	return
}

func (h *GetTaskHandler) validateRequestData(requestData *getTaskRequest) error {
	return validator.New().Struct(requestData)
}

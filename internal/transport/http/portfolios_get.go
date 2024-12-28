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
		err error
	)

	portfoliosList, err := h.getPortfoliosListCommand.GetPortfoliosList(ctx)

	if err != nil {
		GetErrorResponse(w, h.name, fmt.Errorf("command handler failed: %w", err), http.StatusInternalServerError)
		return
	}

	portfoliosJson, err := json.Marshal(portfoliosList)
	if err != nil {
		GetErrorResponse(w, h.name, fmt.Errorf("json marshalling failed: %w", err), http.StatusInternalServerError)
		return
	}

	GetSuccessResponseWithBody(w, portfoliosJson)
}

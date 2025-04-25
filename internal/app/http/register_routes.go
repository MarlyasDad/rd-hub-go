package http

import (
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/index"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http/subscribers"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"net/http"
)

func RegisterHandlers(mux *http.ServeMux, brokerClient *alor.Client) {
	index.RegisterRoutes(mux)
	subscribers.RegisterRoutes(mux, brokerClient)
}

package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	httpmiddlewares "github.com/MarlyasDad/rd-hub-go/internal/app/http/middlewares"
)

type Server struct {
	Mux    *http.ServeMux
	server *http.Server
}

func New(config Config) Server {
	mux := http.NewServeMux()

	httpHandler := httpmiddlewares.RemoveTrailingMiddleware(httpmiddlewares.LoggingMiddleware(httpmiddlewares.CorsMiddleware(mux)))

	return Server{
		Mux: mux,
		server: &http.Server{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port), Handler: httpHandler, ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second},
	}
}

func (s Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v", err)
		}
	}()
}

func (s Server) Stop() {
	_ = s.server.Shutdown(context.Background())
}

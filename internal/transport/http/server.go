package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	httpmiddlewares "github.com/MarlyasDad/rd-hub-go/internal/transport/http/middlewares"
)

type Server struct {
	Mux    *http.ServeMux
	server *http.Server
}

func New(config Config) Server {
	mux := http.NewServeMux()

	return Server{
		Mux:    mux,
		server: &http.Server{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port), Handler: httpmiddlewares.LoggingMiddleware(httpmiddlewares.CorsMiddleware(mux))},
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

//func (s Server) AddHandler(pattern string, handler http.Handler) {
//	s.mux.Handle(pattern, handler)
//}

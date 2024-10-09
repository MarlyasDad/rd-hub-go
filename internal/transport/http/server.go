package http

import (
	"fmt"
	"log"
	"net/http"
)

type HttpServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func New(config Config) HttpServer {
	mux := http.NewServeMux()

	return HttpServer{
		mux:    mux,
		server: &http.Server{Addr: fmt.Sprintf("%s:%d", config.server.Host, config.server.Port), Handler: middlewares.LoggingMiddlewareHandler(middlewares.EnableCors(mux))},
	}
}

func Start() {
	// Start webserver
	log.Println("Starting Webserver")
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v", err)
		}
	}()
}

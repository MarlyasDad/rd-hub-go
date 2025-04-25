package index

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/dist/index.html")
	})
	// Static files
	httpFileServer := http.FileServer(http.Dir("./web/dist/assets"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", httpFileServer))
}

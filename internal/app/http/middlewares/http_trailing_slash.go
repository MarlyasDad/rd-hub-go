package middlewares

import (
	"log"
	"net/http"
	"strings"
)

func RemoveTrailingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") {
			newURI := strings.TrimSuffix(req.URL.Path, "/")
			req.URL.Path = newURI
			req.RequestURI = newURI
			log.Printf("URI: %s %s CHANGED: trailing slash removed", req.Method, req.URL.Path)
		}
		next.ServeHTTP(w, req)
	})
}

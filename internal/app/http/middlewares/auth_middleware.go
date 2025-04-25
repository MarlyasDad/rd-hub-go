package middlewares

import (
	// appHttp "github.com/MarlyasDad/rd-hub-go/internal/transport/http"
	"log"
	"net/http"
	"strings"
)

func HttpAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bearer := req.Header.Get("Authorization")

		if len(bearer) == 0 {
			// appHttp.GetUnauthorizedResponse(w)
			return
		}

		token := strings.Replace(bearer, "Bearer ", "", -1)
		log.Println(token)

		//// Проверить активность токена
		//active, err := keycloakClient.IsTokenActive(token)
		//if err != nil {
		//	appHttp.GetUnauthorizedResponse(w)
		//	return
		//}
		//
		//if !active {
		//	appHttp.GetUnauthorizedResponse(w)
		//	return
		//}

		next.ServeHTTP(w, req)
	})
}

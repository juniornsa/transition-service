package middleware

import (
	"github.com/gorilla/context"
	"log"
	"net/http"
)

func LoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		ch := make(chan string, 1)
		context.Set(r, "ch", ch)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}


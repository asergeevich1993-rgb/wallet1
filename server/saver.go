package server

import (
	"log"
	"net/http"
)

func Saver(router http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Panic", err)
				http.Error(w, "fatal error", http.StatusInternalServerError)
			}
		}()
		router.ServeHTTP(w, r)
	})
}

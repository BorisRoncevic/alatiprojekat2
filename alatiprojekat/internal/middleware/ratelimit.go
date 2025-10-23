package middleware

import (
	"log"
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimitMiddleware(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Println("MIDDLEWARE POZVAN!")
			if !limiter.Allow() {
				log.Println("ZAHTEV BLOKIRAN!")
				http.Error(w, "Previše zahteva", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
